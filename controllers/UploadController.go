package controllers

import (
	"github.com/ipweb-group/file-server/backgroundWorker"
	"github.com/ipweb-group/file-server/controllers/uploadHelper"
	"github.com/ipweb-group/file-server/db/mongodb"
	"github.com/ipweb-group/file-server/putPolicy"
	"github.com/ipweb-group/file-server/putPolicy/mediaHandler"
	"github.com/kataras/iris"
	"mime"
	"path"
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

	// 创建 File 对象
	fileObj := mongodb.File{
		OriginalFilename: fileHeader.Filename,
		FileSize:         fileHeader.Size,
		PutPolicy:        policy,
	}

	// 1. 获取文件的 CID
	lg.Info("Upload client key is ", policy.ClientKey)
	fileObj.Id, err = uploadHelper.GetFileHash(policy.ClientKey, file)
	if err != nil {
		lg.Error("Failed to get file hash, " + err.Error())
		throwError(iris.StatusInternalServerError, "Failed to get file hash, "+err.Error(), ctx)
		return
	}

	// 2. 保存上传后的文件到临时目录下
	fileExt := path.Ext(fileObj.OriginalFilename)
	fileObj.MimeType = mime.TypeByExtension(fileExt)
	tmpFilePath, err := uploadHelper.WriteTmpFile(file, fileObj.Id, fileExt)

	uploadTask := backgroundWorker.UploadTask{
		Hash:          fileObj.Id,
		CacheFilePath: tmpFilePath,
	}

	// 2. 检查数据库中是否存在相同 hash 的记录
	_f, err := mongodb.GetFileByHash(fileObj.Id)
	if err == nil {
		lg.Info("File is already in DB, will return success directly")
		// 文件已经存在时，添加上传任务，并直接回调上传成功
		uploadTask.Enqueue(backgroundWorker.IPFS)

		// 回调上传成功
		uploadHelper.PostUpload(ctx, _f, policy, fileObj.OriginalFilename)
		return
	}

	// 3. 添加上传任务到队列
	uploadTask.Enqueue(backgroundWorker.OSS)
	uploadTask.Enqueue(backgroundWorker.IPFS)

	// 4. 检查文件类型及媒体信息
	// 检测媒体文件信息。当上传文件为图片或视频时，会检测文件的尺寸、时长等信息
	mediaInfo, needCovert, err := mediaHandler.DetectMediaInfo(tmpFilePath, fileObj.MimeType)
	if err == nil {
		fileObj.MediaInfo.Width = mediaInfo.Width
		fileObj.MediaInfo.Height = mediaInfo.Height
		fileObj.MediaInfo.Duration = mediaInfo.Duration
		fileObj.MediaInfo.Type = mediaInfo.Type
	} else {
		lg.Warnf("Detect media info failed, [%v] \n", err)
	}

	if needCovert {
		// TODO ConvertStatus 这个字段好像没什么用
		fileObj.ConvertStatus = mongodb.FileCovertStatusProcessing
	}

	// 写入文件信息到数据库
	err = fileObj.Insert()
	if err != nil {
		lg.Error("Insert file object to db failed, ", err.Error())
		throwError(iris.StatusInternalServerError, "Failed to save file record", ctx)
		return
	}

	// 回调上传成功状态
	uploadHelper.PostUpload(ctx, fileObj, policy, fileObj.OriginalFilename)
	return
}

func throwError(statusCode int, error string, ctx iris.Context) {
	ctx.Application().Logger().Error(error)
	ctx.StatusCode(statusCode)
	_, _ = ctx.JSON(iris.Map{
		"error": error,
	})
}
