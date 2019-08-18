package backgroundWorker

import (
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/golog"
	"time"
)

var lg *golog.Logger

func StartWorker() {
	lg = utils.GetLogger()

	lg.Info("Background worker is started")

	blockFlag := make(chan bool)

	for {
		time.Sleep(2 * time.Second)

		go jobDetector(blockFlag)

		<-blockFlag
	}
}

func jobDetector(flag chan bool) {

	// 检查是否有上传任务
	_uploadTask, err := DequeueUploadTask()
	if err == nil {
		_uploadTask.Upload(flag)
		return
	}

	// 检查是否有下载任务
	_downloadTask, err := DequeueDownloadTask()
	if err == nil {
		_downloadTask.Download(flag)
		return
	}

	flag <- false
}
