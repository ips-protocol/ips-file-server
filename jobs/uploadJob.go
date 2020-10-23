package jobs

import (
	"encoding/json"
	"errors"
	"github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/putPolicy/mediaHandler"
	"github.com/ipweb-group/file-server/utils"
	"strconv"
	"strings"
	"time"
)

type UploadJob struct {
	FileRecordId  string                 `json:"fileRecordId"` // 文件保存在 mongo file_records 表中的 ID
	Hash          string                 `json:"hash"`
	CacheFilePath string                 `json:"cacheFilePath"`
	Filename      string                 `json:"filename"`
	FileSize      int64                  `json:"fileSize"`
	ClientKey     string                 `json:"clientKey"`
	MediaInfo     mediaHandler.MediaInfo `json:"mediaInfo"`
	RetryTimes    int                    `json:"retryTimes"`
}

func (ut *UploadJob) ToJSON() string {
	str, _ := json.Marshal(ut)
	return string(str)
}

// 获取用于保存在 ZSET 中的 member 名称
func (ut *UploadJob) GetMemberName() string {
	return "IPFS:" + ut.FileRecordId
}

// 解码 Member 名称，并返回 Target 和 fileRecordId
func (ut UploadJob) DecodeMemberName(memberName string) (target, fileRecordId string) {
	ret := strings.SplitN(memberName, ":", 2)
	target = ret[0]
	fileRecordId = ret[1]
	return
}

// 添加任务到队列
// target 为 CDN 或 IPFS
func (ut *UploadJob) Enqueue(taskTime int64) {
	redisClient := redisdb.GetClient()
	// 添加任务
	redisClient.Set(GetUploadTaskCacheKey(ut.FileRecordId), ut.ToJSON(), 0)
	// 添加任务到队列
	redisClient.ZAdd(GetUploadQueueCacheKey(), redis.Z{
		Score:  float64(taskTime),
		Member: ut.GetMemberName(),
	})
}

// 从上传队列中弹出一个任务，同时从 ZSET 中移除该任务，并返回解析后的任务内容
func (ut UploadJob) Dequeue() (ret *UploadJob, err error) {
	lg := utils.GetLogger()
	redisClient := redisdb.GetClient()
	_t, err := redisClient.ZRangeByScore(GetUploadQueueCacheKey(), redis.ZRangeBy{
		Min:    "0",
		Max:    strconv.FormatInt(time.Now().Unix(), 10),
		Offset: 0,
		Count:  1,
	}).Result()
	if err != nil {
		return
	}

	if len(_t) == 0 {
		err = errors.New("no available jobs")
		return
	}

	key := _t[0]
	// 用冒号分割 Key（前面的 target 参数已不再需要）
	_, fileRecordId := UploadJob{}.DecodeMemberName(key)

	lg.Info("An upload job detected, ", key)

	// 获取任务缓存
	// 缓存无法匹配，或本地缓存文件不存在时，均要移除 ZSET 中对应的记录
	cacheKey := GetUploadTaskCacheKey(fileRecordId)
	_cacheStr, err := redisClient.Get(cacheKey).Result()
	if err != nil {
		lg.Error("Get cache content failed, ", err.Error())
		ut.removeQueueTask(key)
		return
	}
	if _cacheStr == "" {
		ut.removeQueueTask(key)
		err = errors.New("get upload task content failed")
		return
	}

	lg.Info("Job cache content is ", _cacheStr)

	parsedJob := UploadJob{}
	err = json.Unmarshal([]byte(_cacheStr), &parsedJob)
	if err != nil {
		lg.Error("Parse job content failed, ", err.Error())
		return
	}

	ret = &parsedJob

	// 移除 ZSET 中的任务
	ut.removeQueueTask(key)

	return
}

// 从队列中移除指定的任务
func (ut UploadJob) removeQueueTask(cacheKey string) {
	//redisClient := redisdb.GetClient()
	//redisClient.ZRem(GetUploadQueueCacheKey(), cacheKey)
	//utils.GetLogger().Info("Upload job ZSET cache removed")
}

// 获取任务缓存 Key
func GetUploadTaskCacheKey(fileRecordId string) string {
	return "IPWEB:FS:TASK:UP:" + fileRecordId
}

// 获取上传队列名字 （ZSET）
func GetUploadQueueCacheKey() string {
	return "IPWEB:FS:QUEUE:UP"
}
