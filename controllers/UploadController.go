package controllers

import (
	"github.com/ipweb-group/file-server/controllers/uploadHelper"
	"github.com/ipweb-group/file-server/externals/mongodb/fileRecord"
	"github.com/ipweb-group/file-server/jobs"
	"github.com/ipweb-group/file-server/putPolicy"
	"github.com/ipweb-group/file-server/putPolicy/mediaHandler"
	"github.com/ipweb-group/file-server/uploaders"
	"github.com/kataras/iris"
	"mime"
	"os"
	"path"
	"time"
)

type UploadController struct{}

/**
 * 文件上传
 */
func (s *UploadController) Upload(ctx iris.Context) {
	lg := ctx.Application().Logger()

	// 处理跨域响应
	corsResponse(ctx)
	if ctx.Request().Method == iris.MethodOptions {
		ctx.StatusCode(iris.StatusNoContent)
		_, _ = ctx.WriteString("")
		return
	}

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

	// 强制 endUser 为必填项
	if policy.EndUser == "" {
		throwError(iris.StatusUnprocessableEntity, "EndUser must be provided", ctx)
		return
	}

	// 获取表单上传的文件
	file, fileHeader, err := ctx.FormFile("file")
	if err != nil {
		throwError(iris.StatusBadRequest, "Invalid File", ctx)
		return
	}

	defer file.Close()

	// TODO 上传有效期的校验

	// TODO 文件大小限制

	var hash string
	var mimeType string
	var filename = fileHeader.Filename
	var size = fileHeader.Size
	var mediaInfo mediaHandler.MediaInfo

	// 1. 获取文件的 CID
	lg.Info("Upload client key is ", policy.ClientKey)
	hash, err = uploadHelper.GetFileHash(policy.ClientKey, file)
	if err != nil {
		lg.Error("Failed to get file hash, " + err.Error())
		throwError(iris.StatusInternalServerError, "Failed to get file hash, "+err.Error(), ctx)
		return
	}
	lg.Info("File CID is ", hash)

	// 2. 保存上传后的文件到临时目录下
	fileExt := path.Ext(filename)
	mimeType = mime.TypeByExtension(fileExt)
	tmpFilePath, err := uploadHelper.WriteTmpFile(file, fileExt)

	// 3. 检查文件类型及媒体信息
	// 检测媒体文件信息。当上传文件为图片或视频时，会检测文件的尺寸、时长等信息
	mediaInfo, err = mediaHandler.DetectMediaInfo(tmpFilePath, mimeType)
	if err != nil {
		lg.Warnf("Detect media info failed, [%v]", err)
	}

	// 4. 构建上传记录信息，并写入到数据库
	uploadFileRecord := fileRecord.FileRecord{
		Hash:        hash,
		Client:      decodedPutPolicy.AppClient.Description,
		ClientAppId: decodedPutPolicy.AppClient.AccessKey,
		Filename:    filename,
		MimeType:    mimeType,
		Size:        size,
		PutPolicy:   policy,
		MediaInfo:   mediaInfo,
	}
	uploadFileRecordId, err := uploadFileRecord.Insert()
	if err != nil {
		lg.Error("Insert upload file record to mongodb failed, ", err)
		throwError(iris.StatusInternalServerError, "Failed to save file record", ctx)
		return
	}

	lg.Info("Insert record to DB successful, object ID is ", uploadFileRecordId)

	// 5. 添加上传任务到队列
	uploadJob := jobs.UploadJob{
		FileRecordId:  uploadFileRecordId,
		Hash:          hash,
		CacheFilePath: tmpFilePath,
		Filename:      filename,
		FileSize:      size,
		ClientKey:     policy.ClientKey,
		MediaInfo:     mediaInfo,
	}

	uploadJob.Enqueue(time.Now().Unix())

	// 另开一个线程直接上传到 CDN
	go func() {
		uploader := uploaders.CDNUploader{
			Job: &uploadJob,
		}
		if err := uploader.Upload(); err != nil {
			lg.Error("Uploader to CDN failed, ", uploader.Job.Hash)
		}
		// 上传到 CDN 成功后，直接删除临时文件
		_ = os.Remove(tmpFilePath)
	}()

	// 回调上传成功状态
	uploadHelper.PostUpload(ctx, uploadFileRecord)
	return
}

func throwError(statusCode int, error string, ctx iris.Context) {
	ctx.Application().Logger().Error(error)
	ctx.StatusCode(statusCode)
	_, _ = ctx.JSON(iris.Map{
		"error": error,
	})
}

func corsResponse(ctx iris.Context) {
	origin := ctx.GetHeader("origin")
	if origin != "" {
		ctx.Header("Access-Control-Allow-Origin", origin)
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS")
		ctx.Header("Access-Control-Max-Age", "86400")
	}

	headers := ctx.GetHeader("access-control-request-headers")
	if headers != "" {
		ctx.Header("Access-Control-Allow-Headers", headers)
	}
}
