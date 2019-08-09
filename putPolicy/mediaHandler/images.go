package mediaHandler

import (
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// 获取图片大小
func GetImageInfo(filePath string, mediaInfo *MediaInfo) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return
	}

	defer file.Close()

	config, imageType, err := image.DecodeConfig(file)
	if err != nil {
		return
	}

	mediaInfo.Width = config.Width
	mediaInfo.Height = config.Height
	mediaInfo.Type = imageType
	return
}
