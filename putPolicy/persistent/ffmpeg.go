package persistent

import (
	"fmt"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/utils"
)

func FfmpegCovertVideo(task *Task) (result Result) {
	lg := utils.GetLogger()
	ffmpeg := config.GetConfig().External.Ffmpeg
	inputFilePath := task.FilePath
	outputFilePath := inputFilePath + "-convert.mp4"

	// 判断如果视频是 h264 并且音频为 aac 或 mp3 的话，将不再进行转换，而是直接回调成功
	if task.MediaInfo.Type == "h264" || task.MediaInfo.Type == "h264/aac" || task.MediaInfo.Type == "h264/mp3" {
		lg.Infof("Video is of type %s, no need to be converted", task.MediaInfo.Type)

		result.Desc = fmt.Sprintf("Video is of type %s, no need to be converted", task.MediaInfo.Type)
		result.Code = CodeSuccess
		result.DstHash = task.Cid
		result.outputFilePath = task.FilePath // 重写输出文件路径即为输入文件路径
		return
	}

	command := fmt.Sprintf("%s -stats -y -hide_banner -i %s -c:v libx264 -c:a aac %s", ffmpeg, inputFilePath, outputFilePath)
	lg.Infof("Start ffmpeg converter, command is %s", command)

	_ret, err := utils.ExecCommand(command)
	lg.Info(_ret)
	if err != nil {
		lg.Errorf("Convert task failed [%v] \n", err)
		result.Code = CodeFailed
		result.Desc = err.Error() + _ret
	} else {
		lg.Info("Convert completed")
		result.Code = CodeSuccess
		result.outputFilePath = outputFilePath
	}

	return
}

func FfmpegVideoThumb(task *Task) (result Result) {
	lg := utils.GetLogger()
	ffmpeg := config.GetConfig().External.Ffmpeg
	inputFilePath := task.FilePath
	outputFilePath := inputFilePath + "-convert.jpg"

	command := fmt.Sprintf("%s -stats -y -hide_banner -i %s -ss 1 -frames:v 1 -f image2 %s", ffmpeg, inputFilePath, outputFilePath)
	lg.Info("Start ffmpeg converter, command is ", command)

	_ret, err := utils.ExecCommand(command)
	lg.Info(_ret)
	if err != nil {
		lg.Errorf("Convert task failed [%v] \n", err)
		result.Code = CodeFailed
		result.Desc = err.Error() + _ret

	} else {
		lg.Info("Convert completed")
		result.Code = CodeSuccess
		result.outputFilePath = outputFilePath
	}

	return
}
