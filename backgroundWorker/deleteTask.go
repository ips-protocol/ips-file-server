package backgroundWorker

import (
	"errors"
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/utils"
	"strings"
)

type DeleteTask struct {
	Hash      string `json:"hash"`
	ClientKey string `json:"clientKey"`
}

// 将任务内容序列化为单个字符串，用于保存在 Redis 中
func (dt *DeleteTask) Serialize() string {
	return dt.Hash + ":" + dt.ClientKey
}

// 将从 Redis 中取出的序列化的字符串解析为任务
func (dt *DeleteTask) Deserialize(taskStr string) {
	parts := strings.SplitN(taskStr, ":", 2)
	dt.Hash = parts[0]
	dt.ClientKey = parts[1]
}

// 添加任务到队列
func (dt *DeleteTask) Enqueue() {
	redisdb.GetClient().RPush(GetDeleteQueueCacheKey(), dt.Serialize())
}

// 获取删除队列的名字（LIST）
func GetDeleteQueueCacheKey() string {
	return "IPWEB:FS:QUEUE:DEL"
}

// 执行删除任务
func (dt *DeleteTask) Delete(completed chan bool) {
	// TODO 从 OSS 中删除文件

	// TODO 检查是否包含转换后的文件，如果有，也全部删除

	// TODO 从 IPFS 中删除文件

	completed <- true
}

// 从删除队列中弹出一个任务
func DequeueDeleteTask() (dt DeleteTask, err error) {
	lg := utils.GetLogger()
	redisClient := redisdb.GetClient()

	hash, err := redisClient.LPop(GetDeleteQueueCacheKey()).Result()
	if err != nil {
		return
	}

	if hash == "" {
		err = errors.New("no available task")
		return
	}

	lg.Info("A delete task detected, ", hash)

	dt = DeleteTask{Hash: hash}
	return
}
