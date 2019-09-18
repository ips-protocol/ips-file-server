package backgroundWorker

import (
	"errors"
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/utils"
	"io/ioutil"
)

type DownloadTask struct {
	Hash string `json:"hash"`
}

// 添加任务到队列
func (dt *DownloadTask) Enqueue() {
	redisClient := redisdb.GetClient()
	redisClient.RPush(GetDownloadQueueCacheKey(), dt.Hash)
}

// 执行下载任务
func (dt *DownloadTask) Download(completed chan bool) {
	rpcClient, _ := utils.GetClientInstance()
	_, err := rpcClient.Download(dt.Hash, ioutil.Discard)
	if err != nil {
		lg.Error("Background download failed, ", err)
	}
	completed <- true
}

// 获取下载队列的名字（LIST）
func GetDownloadQueueCacheKey() string {
	return "IPWEB:FS:QUEUE:DOWN"
}

// 从下载队列中弹出一个任务
func DequeueDownloadTask() (dt DownloadTask, err error) {
	lg := utils.GetLogger()
	redisClient := redisdb.GetClient()

	hash, err := redisClient.LPop(GetDownloadQueueCacheKey()).Result()
	if err != nil {
		return
	}

	if hash == "" {
		err = errors.New("no available task")
		return
	}

	lg.Info("A download task detected, ", hash)

	dt = DownloadTask{Hash: hash}
	return
}
