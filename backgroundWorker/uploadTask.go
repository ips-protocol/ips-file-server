package backgroundWorker

import (
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/db/redisdb"
	"time"
)

const (
	OSS  = "OSS"
	IPFS = "IPFS"
)

type UploadTask struct {
	Hash          string `json:"hash"`
	CacheFilePath string `json:"cacheFilePath"`
	RetryTimes    int    `json:"retryTimes"`
}

func (ut *UploadTask) ToJSON() string {
	str, _ := json.Marshal(ut)
	return string(str)
}

// 获取用于保存在 ZSET 中的 member 名称
func (ut *UploadTask) GetMemberName(target string) string {
	return target + ":" + ut.Hash
}

// 添加任务到队列
// target 为 OSS 或 IPFS
func (ut *UploadTask) Enqueue(target string) {
	redisClient := redisdb.GetClient()
	// 添加任务
	redisClient.Set(GetUploadTaskCacheKey(ut), ut.ToJSON(), 0)
	// 添加任务到队列
	redisClient.ZAdd(GetUploadQueueCacheKey(), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: ut.GetMemberName(target),
	})
}

// 获取任务缓存 Key
func GetUploadTaskCacheKey(task *UploadTask) string {
	return "IPWEB:FS:TASK:UP:" + task.Hash
}

// 获取上传队列名字 （ZSET）
func GetUploadQueueCacheKey() string {
	return "IPWEB:FS:QUEUE:UP"
}
