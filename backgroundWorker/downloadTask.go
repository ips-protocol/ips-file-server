package backgroundWorker

import (
	"encoding/json"
	"errors"
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/utils"
	"io/ioutil"
)

type DownloadTask struct {
	Hash string `json:"hash"`
}

func (dt *DownloadTask) ToJSON() string {
	str, _ := json.Marshal(dt)
	return string(str)
}

// 添加任务到队列
func (dt *DownloadTask) Enqueue() {
	redisClient := redisdb.GetClient()
	redisClient.RPush(GetDownloadQueueCacheKey(), dt.Hash)
}

// 执行下载任务
func (dt *DownloadTask) Download(completed chan bool) {
	rpcClient, _ := utils.GetClientInstance()
	_, _ = rpcClient.Download(dt.Hash, ioutil.Discard)
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
