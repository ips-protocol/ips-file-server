package controllers

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/ipweb-group/file-server/backgroundWorker"
	"github.com/ipweb-group/file-server/externals/ossClient"
	"github.com/kataras/iris"
	"io"
	"net/http"
	"regexp"
	"strings"
)

type DownloadController struct{}

func (d *DownloadController) StreamedDownload(ctx iris.Context) {
	lg := ctx.Application().Logger()

	cid := ctx.Params().Get("cid")
	if cid == "" {
		throwError(iris.StatusUnprocessableEntity, "Invalid file hash", ctx)
		return
	}

	lg.Info("New download request")

	// 检查请求中是否包含 Range 头，如果包含，则将 Range 的字节部分解析出来
	rangeHeader := getRangeHeader(ctx)

	// 检查文件是否存在于 OSS 中，如果不存在则直接抛出 404 错误
	ossBucket := ossClient.GetBucket()
	filenameInOSS := "files/" + cid
	var objectMeta http.Header
	var err error

	if rangeHeader == "" {
		objectMeta, err = ossBucket.GetObjectDetailedMeta(filenameInOSS)
	} else {
		objectMeta, err = ossBucket.GetObjectDetailedMeta(filenameInOSS, oss.NormalizedRange(rangeHeader))
	}

	if err != nil {
		throwError(iris.StatusNotFound, "File not found", ctx)
		return
	}

	if rangeHeader != "" {
		ctx.StatusCode(iris.StatusPartialContent)
		ctx.Header("Accept-Ranges", "bytes")
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.Header("Content-Range", objectMeta.Get("Content-Range"))
	}

	ctx.Header("Content-Type", objectMeta.Get("Content-Type"))
	ctx.Header("Content-Length", objectMeta.Get("Content-Length"))
	ctx.Header("Date", objectMeta.Get("Date"))
	ctx.Header("Etag", objectMeta.Get("Etag"))

	lg.Info("Download file from OSS")

	// 从 OSS 上下载指定的文件
	var body io.ReadCloser
	if rangeHeader == "" {
		body, err = ossBucket.GetObject(filenameInOSS)
	} else {
		body, err = ossBucket.GetObject(filenameInOSS, oss.NormalizedRange(rangeHeader))
	}

	if err != nil {
		lg.Error("Get file from OSS failed, ", err)
		throwError(iris.StatusInternalServerError, "Get file failed, "+err.Error(), ctx)
		return
	}
	defer body.Close()

	// 添加后台下载任务。Range 请求时，仅从 0 开始的请求需要后台下载
	defer func() {
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

// 获取请求中的 Range 字节部分
func getRangeHeader(ctx iris.Context) string {
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
