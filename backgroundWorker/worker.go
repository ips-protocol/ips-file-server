package backgroundWorker

import (
	"context"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/golog"
	"time"
)

var lg *golog.Logger

func StartWorker(ctx context.Context) {
	lg = utils.GetLogger()

	lg.Info("Background worker is started")

	blockFlag := make(chan bool, 1)

	go jobDetector(blockFlag)

	for {
		select {
		case <-ctx.Done():
			// FIXME 平滑关闭功能尚有 Bug
			lg.Info("Background worker is canceling")
			<-blockFlag
			// ctx 已取消
			return

		case <-blockFlag:
			time.Sleep(2 * time.Second)
			go jobDetector(blockFlag)
		}
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

	// 检查是否有删除任务
	//_deleteTask, err := DequeueDeleteTask()
	//if err == nil {
	//	_deleteTask.Delete(flag)
	//	return
	//}

	flag <- false
}
