package controllers

import (
	"github.com/ipweb-group/file-server/putPolicy"
	"github.com/ipweb-group/file-server/putPolicy/mediaHandler"
	"github.com/ipweb-group/file-server/putPolicy/persistent"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/iris"
	"mime"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

type UploadController struct{}

/**
 * 文件上传
 */
func (s *UploadController) Upload(ctx iris.Context) {
	lg := ctx.Application().Logger()
	token := ctx.FormValue("token")
	if len(token) == 0 {
		throwError(iris.StatusUnprocessableEntity, "No Upload Token Specified", ctx)
		return
	}

	// 解码上传 Token
	decodedPutPolicy, err := putPolicy.DecodePutPolicyString(token)
	if err != nil {
		throwError(iris.StatusInternalServerError, err.Error(), ctx)
		return
	}
	policy := decodedPutPolicy.PutPolicy

	// 获取表单上传的文件
	file, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		throwError(iris.StatusBadRequest, "Invalid File", ctx)
		return
	}

	defer file.Close()

	// TODO 上传有效期的校验

	// TODO 文件大小限制

	// 上传文件到 IPFS
	// 根据策略中有没有传 client key 来决定是使用私有账户上传还是使用公共账户上传
	rpcClient, _ := utils.GetClientInstance()
	var cid string
	lg.Info("Upload client key is ", policy.ClientKey)
	if policy.ClientKey != "" {
		cid, err = rpcClient.UploadByClientKey(policy.ClientKey, file, fileHeader.Filename, fileHeader.Size)
	} else {
		cid, err = rpcClient.Upload(file, fileHeader.Filename, fileHeader.Size)
	}
	if err != nil {
		throwError(iris.StatusInternalServerError, "Failed to Upload, "+err.Error(), ctx)
		return
	}

	// 写入临时文件
	fileExt := path.Ext(fileHeader.Filename)
	mimeType := mime.TypeByExtension(fileExt)
	tmpFilePath, err := mediaHandler.WriteTmpFile(file, cid, fileExt)

	// 初始化魔法变量对象
	magicVariable := putPolicy.MagicVariable{
		FName:    fileHeader.Filename,
		Hash:     cid,
		FSize:    fileHeader.Size,
		EndUser:  policy.EndUser,
		MimeType: mimeType,
	}

	// 检测媒体文件信息。当上传文件为图片或视频时，会检测文件的尺寸、时长等信息
	mediaInfo, err := mediaHandler.DetectMediaInfo(tmpFilePath, mimeType)
	if err == nil {
		magicVariable.Width = mediaInfo.Width
		magicVariable.Height = mediaInfo.Height
		magicVariable.Duration = mediaInfo.Duration
	} else {
		lg.Warnf("Detect media info failed, [%v] \n", err)
	}

	// 处理持久化任务
	persistentTask := persistent.Task{
		Cid:                 cid,
		FilePath:            tmpFilePath,
		PersistentOps:       policy.PersistentOps,
		PersistentNotifyUrl: policy.PersistentNotifyUrl,
		MediaInfo:           mediaInfo,
		ClientKey:           policy.ClientKey,
	}
	shouldRemoveTmpFile := persistentTask.CheckShouldQueueTask()

	if !shouldRemoveTmpFile {
		// 如果需要添加到队列中，则在当前函数结束时将任务添加到队列，以避免队列过早执行
		defer persistentTask.Queue()
	}

	// 删除临时文件
	if shouldRemoveTmpFile {
		_ = os.Remove(tmpFilePath)
	}

	// 如果上传策略中指定了 returnBody，就去解析这个 returnBody。如果同时指定了 returnUrl，将会 303 跳转到该地址，
	// 否则就直接将 returnBody 的内容显示在浏览器上
	lg.Debug("Return body is ", policy.ReturnBody)
	lg.Debug("Return Url is ", policy.ReturnUrl)
	if policy.ReturnBody != "" || policy.ReturnUrl != "" {
		returnBody := magicVariable.ApplyMagicVariables(policy.ReturnBody, putPolicy.EscapeJSON)

		lg.Debug("Return body with magic variables: ", returnBody)

		// 当设置了 ReturnUrl 时，将会跳转到指定的地址
		if match, _ := regexp.MatchString("(?i)^https?://", policy.ReturnUrl); policy.ReturnUrl != "" && match {
			var l string
			if strings.Contains(policy.ReturnUrl, "?") {
				l = "&"
			} else {
				l = "?"
			}
			redirectUrl := policy.ReturnUrl + l + "upload_ret=" + url.QueryEscape(returnBody)
			lg.Info("Redirect to URL ", redirectUrl)

			ctx.Redirect(redirectUrl, iris.StatusSeeOther)
			return
		}

		// 未设置 returnUrl 时，直接返回 returnBody 的内容
		lg.Info("No returnUrl specified or URL is invalid, will show return body content: ", returnBody)
		ctx.Header("Content-Type", "application/json; charset=UTF-8")
		_, _ = ctx.WriteString(returnBody)
		return
	}

	// 如果上传策略中指定了回调地址，就异步去请求该地址
	if policy.CallbackUrl != "" {
		responseBody, err := policy.ExecCallback(magicVariable, putPolicy.EscapeURL)
		if err != nil {
			lg.Debugf("Callback to %s failed, %v \n", policy.CallbackUrl, err)
			throwError(utils.StatusCallbackFailed, "Callback Failed, "+err.Error(), ctx)
			return
		}
		lg.Debugf("Callback to %s responds %s \n", policy.CallbackUrl, responseBody)

		ctx.Header("Content-Type", "application/json; charset=UTF-8")
		_, _ = ctx.WriteString(responseBody)
		return
	}

	// 未指定回调地址时，返回默认内容
	_, _ = ctx.JSON(iris.Map{
		"hash":   cid,
		"length": fileHeader.Size,
	})
}

func throwError(statusCode int, error string, ctx iris.Context) {
	ctx.Application().Logger().Error(error)
	ctx.StatusCode(statusCode)
	_, _ = ctx.JSON(iris.Map{
		"error": error,
	})
}
