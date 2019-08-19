package backgroundWorker

import (
	"encoding/json"
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/externals/aliyun"
	"github.com/ipweb-group/file-server/externals/ossClient"
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/utils"

	"mime"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	CDN  = "CDN"
	IPFS = "IPFS"
)

type UploadTask struct {
	Hash          string `json:"hash"`
	CacheFilePath string `json:"cacheFilePath"`
	Filename      string `json:"filename"`
	FileSize      int64  `json:"fileSize"`
	ClientKey     string `json:"clientKey"`
	RetryTimes    int    `json:"retryTimes"`
}

type UploadTaskWithTarget struct {
	Target     string     `json:"target"`
	UploadTask UploadTask `json:"uploadTask"`
}

func (ut *UploadTask) ToJSON() string {
	str, _ := json.Marshal(ut)
	return string(str)
}

// 获取用于保存在 ZSET 中的 member 名称
func (ut *UploadTask) GetMemberName(target string) string {
	return target + ":" + ut.Hash
}

// 解码 Member 名称，并返回 Target 和 Hash
func (ut UploadTask) DecodeMemberName(memberName string) (target, hash string) {
	ret := strings.SplitN(memberName, ":", 2)
	target = ret[0]
	hash = ret[1]
	return
}

// 上传文件到 IPFS
func (ut *UploadTask) UploadToIPFS() (err error) {
	lg := utils.GetLogger()
	rpcClient, _ := utils.GetClientInstance()
	var cid string

	lg.Infof("Prepare to upload to IPFS, cid is %s, client key is %s", ut.Hash, ut.ClientKey)
	if ut.ClientKey == "" {
		cid, err = rpcClient.UploadWithPath(ut.CacheFilePath)
	} else {
		cid, err = rpcClient.UploadWithPathByClientKey(ut.ClientKey, ut.CacheFilePath)
	}

	if err != nil {
		return
	}

	utils.GetLogger().Infof("Upload to IPFS successful. Current cid is %s, target cid is %s", ut.Hash, cid)
	return
}

// 上传文件到 OSS
func (ut *UploadTask) UploadToOSS() (err error) {
	mimeType := mime.TypeByExtension(path.Ext(ut.CacheFilePath))
	bucket := ossClient.GetBucket()
	ossFilePath := "files/" + ut.Hash
	err = bucket.PutObjectFromFile(ossFilePath, ut.CacheFilePath, oss.ContentType(mimeType))
	if err != nil {
		return
	}
	lg.Info("Upload to OSS completed")

	// 如果文件是视频类型，同时启动转码服务
	if match, _ := regexp.MatchString("video/.*", mimeType); match {
		lg.Info("File is of type video, will request converting")
		go aliyun.VideoSnapShot(ossFilePath, "converted/"+ut.Hash+"/snapshot.jpg")
		go aliyun.VideoCovert(ossFilePath, "converted/"+ut.Hash+"/playable.mp4")
	}
	return
}

// 添加任务到队列
// target 为 CDN 或 IPFS
func (ut *UploadTask) Enqueue(target string) {
	redisClient := redisdb.GetClient()
	// 添加任务
	redisClient.Set(GetUploadTaskCacheKey(ut.Hash), ut.ToJSON(), 0)
	// 添加任务到队列
	redisClient.ZAdd(GetUploadQueueCacheKey(), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: ut.GetMemberName(target),
	})
}

// 执行上传任务
func (utwt *UploadTaskWithTarget) Upload(completed chan bool) {
	var err error
	lg := utils.GetLogger()
	lg.Info("Begin upload to ", utwt.Target)

	if utwt.Target == CDN {
		err = utwt.UploadTask.UploadToOSS()
	}

	if utwt.Target == IPFS {
		err = utwt.UploadTask.UploadToIPFS()
	}

	if err != nil {
		lg.Warn("Upload failed, ", err.Error())

		// 出错后，重试次数加一
		utwt.UploadTask.RetryTimes++
		// TODO 重新保存回 Redis

	}

	// 上传完成后，检查 Queue 中是否存在相同 hash 的不同类型的任务，如果没有别的任务，
	// 这时就可以安全地删除 Redis 缓存及临时文件
	if !utwt.HasAnotherTask() {
		redisClient := redisdb.GetClient()
		redisClient.Del(GetUploadTaskCacheKey(utwt.UploadTask.Hash))
		_ = os.Remove(utwt.UploadTask.CacheFilePath)
		lg.Infof("File %s has no other upload task, will remove temp file and redis cache", utwt.UploadTask.Hash)
	}

	completed <- true
	return
}

// 判断当前 hash 的文件是否还有其他未完成的上传任务
func (utwt *UploadTaskWithTarget) HasAnotherTask() bool {
	redisClient := redisdb.GetClient()
	hasAnotherTask := false
	for _, target := range []string{CDN, IPFS} {
		member := utwt.UploadTask.GetMemberName(target)
		score, err := redisClient.ZScore(GetUploadQueueCacheKey(), member).Result()
		if err == nil && score > 0 {
			hasAnotherTask = true
		}
	}
	return hasAnotherTask
}

// 获取任务缓存 Key
func GetUploadTaskCacheKey(hash string) string {
	return "IPWEB:FS:TASK:UP:" + hash
}

// 获取上传队列名字 （ZSET）
func GetUploadQueueCacheKey() string {
	return "IPWEB:FS:QUEUE:UP"
}

// 从上传队列中弹出一个任务，同时从 ZSET 中移除该任务，并返回解析后的任务内容
func DequeueUploadTask() (ut UploadTaskWithTarget, err error) {
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
		err = errors.New("no available task")
		return
	}

	key := _t[0]
	// 用冒号分割 Key
	target, hash := UploadTask{}.DecodeMemberName(key)

	lg.Info("An upload task detected, ", key)

	// 获取任务缓存
	cacheKey := GetUploadTaskCacheKey(hash)
	_cacheStr, err := redisClient.Get(cacheKey).Result()
	if err != nil {
		return
	}
	if _cacheStr == "" {
		err = errors.New("get upload task content failed")
		return
	}

	var _ut UploadTask
	err = json.Unmarshal([]byte(_cacheStr), &_ut)
	if err != nil {
		return
	}

	ut = UploadTaskWithTarget{
		Target:     target,
		UploadTask: _ut,
	}

	// 移除 ZSET 中的任务
	redisClient.ZRem(GetUploadQueueCacheKey(), key)

	return
}
