package controllers

import (
	"fmt"
	"github.com/ipweb-group/file-server/fileCache"
	"github.com/kataras/iris"
	"io"
	"os"
	"regexp"
	"strconv"
)

type DownloadController struct{}

func (d *DownloadController) StreamedDownload(ctx iris.Context) {
	lg := ctx.Application().Logger()

	cid := ctx.Params().Get("cid")
	if cid == "" {
		throwError(iris.StatusUnprocessableEntity, "Invalid file hash", ctx)
		return
	}

	// 标识文件是否是从 IPFS 中下载。如果文件是来自缓存而非下载自 IPFS，将会在
	// 请求结束后在后台自动下载该文件以达到计数的目的
	isFileDownloadedFromIPFS := false

	// 检查文件是否存在于缓存中，如果不存在，则将其下载到缓存，并写到 Redis
	if !fileCache.IsCacheAvailable(cid) {
		lg.Info("File cache not available, start downloading")
		// 尝试删除可能存在的缓存文件
		fileCache.RemoveCachedFileAndRedisKey(cid)

		// 从 IPFS 中下载文件
		err := fileCache.DownloadFileToCache(cid)
		if err != nil {
			throwError(iris.StatusInternalServerError, err.Error(), ctx)
			return
		}

		isFileDownloadedFromIPFS = true
	}

	file, fileInfo, err := fileCache.GetCachedFile(cid)
	if err != nil {
		lg.Warnf("Open cache file failed, %s, %v", cid, err)
		throwError(iris.StatusInternalServerError, "Open file failed", ctx)
		return
	}
	defer file.Close()

	lg.Info("Read file from cache, ", cid)

	// 更新文件在缓存中的最后访问时间（该时间用于清理缓存）
	defer func() {
		go fileCache.UpdateFileAccessTimeToNow(cid)
	}()

	// 处理 Range 请求
	rangeHeader := ctx.Request().Header.Get("Range")
	if rangeHeader != "" {
		if match, _ := regexp.MatchString("(?i)bytes=(\\d*)-(\\d*)", rangeHeader); match {
			isRangeStart := handleRangeRequest(ctx, file, fileInfo, rangeHeader)
			// Range 请求为起点时，在后台下载 IPFS 文件
			if isRangeStart {
				go fileCache.BackgroundDownload(cid)
			}
			return
		}
	}

	// 文件并非请求自 IPFS 时，启动后台下载
	if !isFileDownloadedFromIPFS {
		defer func() {
			go fileCache.BackgroundDownload(cid)
		}()
	}

	// 非 Range 请求时，返回文件内容
	ctx.Header("Content-Type", fileInfo.MimeType)

	_, err = io.Copy(ctx.ResponseWriter(), file)
	if err != nil {
		lg.Errorf("Copy file stream to context failed, %v", err)
		throwError(iris.StatusInternalServerError, "Send file content failed", ctx)
		return
	}
}

// 处理 Range 请求，并完成响应。
// 返回此次请求是否一个新的 Range 起点请求的标识（新的起点请求需要后台下载 IPFS 文件）
func handleRangeRequest(ctx iris.Context, file *os.File, fileInfo fileCache.CachedFile, rangeHeader string) (isRangeStart bool) {
	lg := ctx.Application().Logger()
	isRangeStart = false

	lg.Info("Range request detected, will provide range response for ", rangeHeader)

	reg, err := regexp.Compile("(?i)bytes=(\\d*)-(\\d*)")
	if err != nil {
		lg.Error(err)
		return
	}

	matches := reg.FindStringSubmatch(rangeHeader)
	if matches == nil {
		throwError(iris.StatusUnprocessableEntity, "Invalid ranger header", ctx)
		return
	}

	var start, end int64

	// 处理 range 为不同值时截取的起点和终点情况。暂不支持多个 range 的截取
	if matches[1] != "" {
		start, _ = strconv.ParseInt(matches[1], 10, 64)
	}
	if matches[2] != "" {
		end, _ = strconv.ParseInt(matches[2], 10, 64)
	}
	if matches[1] != "" && matches[2] == "" {
		end = fileInfo.Size - 1
	}
	if matches[1] == "" && matches[2] != "" {
		start = fileInfo.Size - end - 1
		end = fileInfo.Size - 1
	}

	if start > fileInfo.Size {
		throwError(iris.StatusRequestedRangeNotSatisfiable, "Range Not Satisfiable", ctx)
		return
	}

	if end >= fileInfo.Size {
		end = fileInfo.Size - 1
	}

	isRangeStart = start == 0

	// 返回指定的文件内容
	length := end - start + 1
	bytes := make([]byte, length)

	_, err = file.ReadAt(bytes, start)
	if err != nil {
		lg.Error(err)
	}

	ctx.StatusCode(iris.StatusPartialContent)
	ctx.Header("Content-Type", fileInfo.MimeType)
	ctx.Header("Content-Length", strconv.FormatInt(length, 10))
	ctx.Header("Accept-Ranges", "bytes")
	ctx.Header("Content-Transfer-Encoding", "binary")
	ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size))

	_, _ = ctx.Write(bytes)
	return
}
