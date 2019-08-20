package controllers

import (
	"github.com/ipweb-group/file-server/backgroundWorker"
	"github.com/ipweb-group/file-server/controllers/uploadHelper"
	"github.com/ipweb-group/file-server/externals/mongodb/fileRecord"
	"github.com/ipweb-group/file-server/putPolicy"
	"github.com/ipweb-group/file-server/putPolicy/mediaHandler"
	"github.com/kataras/iris"
	"mime"
	"path"
	"time"
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

	// 2. 保存上传后的文件到临时目录下
	fileExt := path.Ext(filename)
	mimeType = mime.TypeByExtension(fileExt)
	tmpFilePath, err := uploadHelper.WriteTmpFile(file, hash, fileExt)

	// 3. 检查文件类型及媒体信息
	// 检测媒体文件信息。当上传文件为图片或视频时，会检测文件的尺寸、时长等信息
	mediaInfo, err = mediaHandler.DetectMediaInfo(tmpFilePath, mimeType)
	if err != nil {
		lg.Warnf("Detect media info failed, [%v]", err)
	}

	// 4. 构建上传记录信息，并写入到数据库
	uploadFileRecord := fileRecord.FileRecord{
		Hash:      hash,
		Client:    decodedPutPolicy.AppClient.Description,
		Filename:  filename,
		MimeType:  mimeType,
		Size:      size,
		PutPolicy: policy,
		MediaInfo: mediaInfo,
	}
	uploadFileRecordId, err := uploadFileRecord.Insert()
	if err != nil {
		lg.Error("Insert upload file record to mongodb failed, ", err)
		throwError(iris.StatusInternalServerError, "Failed to save file record", ctx)
		return
	}

	// 5. 添加上传任务到队列
	uploadTask := backgroundWorker.UploadTask{
		FileRecordId:  uploadFileRecordId,
		Hash:          hash,
		CacheFilePath: tmpFilePath,
		Filename:      filename,
		FileSize:      size,
		ClientKey:     policy.ClientKey,
	}

	uploadTask.Enqueue(backgroundWorker.CDN, time.Now().Unix())
	uploadTask.Enqueue(backgroundWorker.IPFS, time.Now().Unix())

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
