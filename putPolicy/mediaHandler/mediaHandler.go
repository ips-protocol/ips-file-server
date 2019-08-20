package mediaHandler

import (
	"github.com/ipweb-group/file-server/utils"
	"regexp"
)

type MediaInfo struct {
	Width    int    `json:"width" bson:"width"`
	Height   int    `json:"height" bson:"height"`
	Duration string `json:"duration" bson:"duration"` // 时长是个浮点数，这里直接用字符串保存
	Type     string `json:"type" bson:"type"`         // 类型。针对图片可能是 jpeg/png/gif；针对视频可能是 h264 等
}

func DetectMediaInfo(filePath string, mimeType string) (info MediaInfo, err error) {
	lg := utils.GetLogger()

	// 如果文件是支持的图片类型，就调用图片处理器获取图片尺寸信息
	if mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/gif" {
		lg.Info("File is of supported image type, will process image size detector")
		err = GetImageInfo(filePath, &info)
		if err != nil {
			return
		}
	}

	// 如果文件是视频类型，就调用视频处理器获取视频的基本信息
	if match, _ := regexp.MatchString("video/.*", mimeType); match {
		lg.Info("File is of type video, will process video information detector")
		err = GetVideoInfo(filePath, &info)
		if err != nil {
			return
		}
	}

	return
}
