package worker

import (
	"github.com/ipweb-group/file-server/jobs"
	"github.com/ipweb-group/file-server/uploaders"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/golog"
	"time"
)

var lg *golog.Logger

func StartWorker() {
	lg = utils.GetLogger()

	lg.Info("Background worker is started")

	blockFlag := make(chan bool, 1)

	go jobDetector(blockFlag)

	for {
		select {

		case <-blockFlag:
			time.Sleep(2 * time.Second)
			go jobDetector(blockFlag)
		}
	}
}

func jobDetector(flag chan bool) {

	// 检查是否有上传任务
	_uploadJob, err := jobs.UploadJob{}.Dequeue()
	//if err != nil {
	//	fmt.Println(err.Error())
	//	flag <- false
	//	return
	//}
	if err == nil {
		uploader := uploaders.IPFSUploader{
			Job: _uploadJob,
		}
		_ = uploader.Upload()
		flag <- true
		return
	}

	// 检查是否有下载任务
	//_downloadTask, err := DequeueDownloadTask()
	//if err == nil {
	//	_downloadTask.Download(flag)
	//	return
	//}

	// 检查是否有删除任务
	//_deleteTask, err := DequeueDeleteTask()
	//if err == nil {
	//	_deleteTask.Delete(flag)
	//	return
	//}

	//flag <- false
}
