package mediaHandler

import (
	"encoding/json"
	"fmt"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/utils"
)

// ffprobe 返回的视频格式信息
type videoInfo struct {
	Streams []videoStream `json:"streams"`
}

// ffprobe 返回的音视频流数据信息
type videoStream struct {
	Index     int    `json:"index"`
	CodecName string `json:"codec_name"`
	CodecType string `json:"codec_type"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Duration  string `json:"duration"`
}

// 获取视频信息
func GetVideoInfo(filePath string, mediaInfo *MediaInfo) (err error) {
	lg := utils.GetLogger()

	// 调用 ffprobe 获取视频信息
	ffprobe := config.GetConfig().External.Ffprobe
	fields := "stream=index,codec_name,codec_type,width,height,duration"
	command := fmt.Sprintf("%s -hide_banner -v quiet -print_format json -show_entries %s  -i %s", ffprobe, fields, filePath)

	result, err := utils.ExecCommand(command)
	if err != nil {
		lg.Warnf("Get video properties failed [%v]", err)
		return
	}

	info := videoInfo{}
	err = json.Unmarshal([]byte(result), &info)
	if err != nil {
		lg.Warnf("Get video properties failed, parse result failed, [%v]", err)
		return
	}

	// 解析返回的流信息
	var videoCodec, audioCodec string

	for _, streamInfo := range info.Streams {
		if videoCodec == "" && streamInfo.CodecType == "video" {
			videoCodec = streamInfo.CodecName

			mediaInfo.Width = streamInfo.Width
			mediaInfo.Height = streamInfo.Height
			mediaInfo.Duration = streamInfo.Duration
		}

		if audioCodec == "" && streamInfo.CodecType == "audio" {
			// 在音频编码前添加左斜线，避免在最终拼接 Type 时由于音频编码为空导致 Type 以左斜线结尾
			audioCodec = "/" + streamInfo.CodecName
		}
	}
	mediaInfo.Type = videoCodec + audioCodec

	return
}
