package controllers

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/ipweb-group/file-server/backgroundWorker"
	"github.com/ipweb-group/file-server/externals/mongodb/fileRecord"
	"github.com/ipweb-group/file-server/externals/ossClient"
	"github.com/kataras/iris"
	"io"
	"regexp"
	"strings"
)

type DownloadController struct{}

var operationMap = map[string]string{
	"playable": "playable.mp4",
	"snapshot": "snapshot.jpg",
}

func (d *DownloadController) StreamedDownload(ctx iris.Context) {
	lg := ctx.Application().Logger()

	cid := ctx.Params().Get("cid")
	operation := ctx.Params().Get("operation")
	if cid == "" {
		throwError(iris.StatusUnprocessableEntity, "Invalid file hash", ctx)
		return
	}

	lg.Infof("New download request, cid is %s, operation is %s", cid, operation)

	// 根据文件 Hash 从数据库中获取文件记录
	file, err := fileRecord.GetFileRecordByHash(cid)
	if err != nil {
		throwError(iris.StatusNotFound, "File not found", ctx)
		return
	}

	targetOssFilePath, targetOssFileProcess := d.parseTargetOssFilePath(file, operation)
	lg.Info("Target OSS file path is ", targetOssFilePath)

	// 检查请求中是否包含 Range 头，如果包含，则将 Range 的字节部分解析出来
	rangeHeader := d.getRangeHeader(ctx)

	ossBucket := ossClient.GetBucket()

	// 计算并处理 OSS 请求的选项
	var ossOptions []oss.Option
	if rangeHeader != "" {
		ossOptions = append(ossOptions, oss.NormalizedRange(rangeHeader))
	}
	if targetOssFileProcess != "" {
		ossOptions = append(ossOptions, oss.Process(targetOssFileProcess))
	}

	// 检查文件是否存在于 OSS 中，如果不存在则直接抛出 404 错误
	_, err = ossBucket.IsObjectExist(targetOssFilePath, ossOptions...)
	if err != nil {
		throwError(iris.StatusNotFound, "File not found", ctx)
		return
	}

	lg.Info("Download file from OSS")

	// 从 OSS 上下载指定的文件
	body, err := ossClient.GetObjectToResponse(ossBucket, targetOssFilePath, ossOptions...)
	if err != nil {
		lg.Error("Get file from OSS failed, ", err)
		throwError(iris.StatusInternalServerError, "Get file failed, "+err.Error(), ctx)
		return
	}
	defer body.Close()

	ctx.StatusCode(body.StatusCode)
	if rangeHeader != "" {
		ctx.StatusCode(iris.StatusPartialContent)
		ctx.Header("Accept-Ranges", "bytes")
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.Header("Content-Range", body.Headers.Get("Content-Range"))
	}
	ctx.Header("Content-Length", body.Headers.Get("Content-Length"))
	ctx.Header("Content-Type", body.Headers.Get("Content-Type"))
	ctx.Header("Date", body.Headers.Get("Date"))
	ctx.Header("Etag", body.Headers.Get("Etag"))

	// 添加后台下载任务。Range 请求时，仅从 0 开始的请求需要后台下载
	defer func() {
		// 如果请求操作为缩略图或视频截图，将不产生下载任务
		if operation == "snapshot" || strings.HasPrefix(operation, "thumb-") {
			return
		}

		if rangeHeader == "" || strings.HasPrefix(rangeHeader, "0-") {
			downloadTask := backgroundWorker.DownloadTask{Hash: cid}
			downloadTask.Enqueue()
			lg.Info("Added background download task to queue")
		}
	}()

	_, err = io.Copy(ctx.ResponseWriter(), body)
	if err != nil {
		lg.Errorf("Copy file stream to context failed, %v", err)
		throwError(iris.StatusInternalServerError, "Send file content failed", ctx)
		return
	}
}

// 解析文件类型和 operation 类型，返回实际可用于返回的 OSS 文件路径
func (d DownloadController) parseTargetOssFilePath(file fileRecord.FileRecord, operation string) (string, string) {
	hash := file.Hash
	originalOssFilePath := "files/" + hash

	// 没有指定 operation 时，直接返回原文件路径
	if operation == "" {
		return originalOssFilePath, ""
	}

	// 指定 Operation 在预置列表中存在时，直接返回对应的转换后的文件
	_, ok := operationMap[operation]
	if ok {
		return d.getConvertedFilePath(hash, operation), ""
	}

	// Operation 以 thumb- 开头
	if strings.HasPrefix(operation, "thumb-") {
		process := "style/" + strings.TrimPrefix(operation, "thumb-")
		var imageFilePath string

		// 判断文件是否是视频，如果是视频就使用其对应的 snapshot 地址
		if match, _ := regexp.MatchString("video/.*", file.MimeType); match {
			imageFilePath = d.getConvertedFilePath(hash, "snapshot")
		} else {
			imageFilePath = originalOssFilePath
		}

		return imageFilePath, process
	}

	return originalOssFilePath, ""
}

// 根据 Operation 获取转换后的文件地址
func (d DownloadController) getConvertedFilePath(hash, operation string) string {
	filename := operationMap[operation]
	return "converted/" + hash + "/" + filename
}

// 获取请求中的 Range 字节部分
func (d DownloadController) getRangeHeader(ctx iris.Context) string {
	rangeHeader := ctx.Request().Header.Get("Range")
	if rangeHeader == "" {
		return ""
	}

	reg, err := regexp.Compile("(?i)bytes=(.*)")
	if err != nil {
		return ""
	}

	matches := reg.FindStringSubmatch(rangeHeader)
	if matches == nil {
		return ""
	}

	return matches[1]
}
