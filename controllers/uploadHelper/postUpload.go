package uploadHelper

import (
	"github.com/ipweb-group/file-server/externals/mongodb/fileRecord"
	"github.com/ipweb-group/file-server/putPolicy"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/iris"
	"net/url"
	"regexp"
	"strings"
)

// 根据文件表中的内容生成魔法变量
func MakeMagicVariable(file fileRecord.FileRecord) putPolicy.MagicVariable {
	return putPolicy.MagicVariable{
		FName:    file.Filename,
		FSize:    file.Size,
		MimeType: file.MimeType,
		EndUser:  file.PutPolicy.EndUser,
		Hash:     file.Hash,
		Width:    file.MediaInfo.Width,
		Height:   file.MediaInfo.Height,
		Duration: file.MediaInfo.Duration,
	}
}

func PostUpload(ctx iris.Context, file fileRecord.FileRecord) {
	lg := utils.GetLogger()
	policy := file.PutPolicy

	// 如果上传策略中指定了 returnBody，就去解析这个 returnBody。如果同时指定了 returnUrl，将会 303 跳转到该地址，
	// 否则就直接将 returnBody 的内容显示在浏览器上
	lg.Debug("Return body is ", policy.ReturnBody)
	lg.Debug("Return Url is ", policy.ReturnUrl)

	// 初始化魔法变量对象
	magicVariable := MakeMagicVariable(file)

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
		"hash":   file.Hash,
		"length": file.Size,
	})
}

func throwError(statusCode int, error string, ctx iris.Context) {
	ctx.Application().Logger().Error(error)
	ctx.StatusCode(statusCode)
	_, _ = ctx.JSON(iris.Map{
		"error": error,
	})
}
