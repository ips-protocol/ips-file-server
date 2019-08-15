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
		go jobDetector(blockFlag)

		<-blockFlag
	}
}

func jobDetector(flag chan bool) {

	hasJob := false
	if !hasJob {
		time.Sleep(2 * time.Second)
		flag <- true
		return
	}

	// 有任务 时执行任务
	return
}
