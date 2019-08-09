package fileCache

import (
	"encoding/json"
	_redis "github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/db/redis"
	"github.com/ipweb-group/file-server/utils"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"
)

const (
	// 文件 Key 保存在 Redis 中的 Key 前缀
	FileKeyPrefix = "IPWEB:FILE:"

	// 缓存文件的 ZSET 列表的 Key
	FileCacheSetKey = "IPWEB:CACHED_FILES"
)

type CachedFile struct {
	Hash     string `json:"hash"`
	Name     string `json:"name"`
	Size     int64  `json:"size"`
	MimeType string `json:"mimeType"`
}

// 获取缓存文件的绝对路径
func GetCacheFilePath(cid string) string {
	return filepath.Join(utils.GetCacheDir(), cid)
}

// 缓存是否存在并可用
// 仅在缓存同时存在于 Redis 和缓存目录时，才认为其可用
func IsCacheAvailable(cid string) bool {
	exist, _ := redis.GetClient().Exists(FileKeyPrefix + cid).Result()
	if exist == int64(0) {
		return false
	}

	cachePath := GetCacheFilePath(cid)
	return utils.PathExists(cachePath)
}

// 后台下载文件，下载后的内容不做任何处理
func BackgroundDownload(cid string) {
	lg := utils.GetLogger()
	lg.Info("Background downloading is started")
	rpcClient, _ := utils.GetClientInstance()
	stream, _, err := rpcClient.StreamRead(cid)
	if err != nil {
		lg.Warn("Background download failed, ", err)
		return
	}
	_ = stream.Close()
}

// 获取缓存文件（不会检查文件是否存在）
func GetCachedFile(cid string) (file *os.File, fileInfo CachedFile, err error) {
	cacheFilePath := GetCacheFilePath(cid)
	lg := utils.GetLogger()

	// 读取 Redis 中的文件信息
	redisClient := redis.GetClient()
	_info, err := redisClient.Get(FileKeyPrefix + cid).Bytes()
	if err == _redis.Nil || err != nil {
		lg.Warn("File key not exists, will remove cache file to keep sync")
		RemoveCachedFileAndRedisKey(cid)
		return
	}

	err = json.Unmarshal(_info, &fileInfo)
	if err != nil {
		return
	}

	// 读取文件
	file, err = os.Open(cacheFilePath)
	if err != nil {
		// 打开缓存文件失败时，将会删除该缓存文件以及对应的缓存
		RemoveCachedFileAndRedisKey(cid)
	}
	return
}

// 添加缓存文件到 Redis 中
func AddCachedFileToRedis(cid string, c CachedFile) {
	str, _ := json.Marshal(c)

	redisClient := redis.GetClient()
	redisClient.Set(FileKeyPrefix+cid, str, 0)

	// 添加文件到 ZSET 列表
	now := time.Now().Unix()
	redisClient.ZAdd(FileCacheSetKey, _redis.Z{
		Score:  float64(now),
		Member: cid,
	})
}

// 更新文件在缓存中的最后访问时间（该时间用于清理缓存）
func UpdateFileAccessTimeToNow(cid string) {
	redis.GetClient().ZAdd(FileCacheSetKey, _redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: cid,
	})
}

// 删除缓存文件，并删除 Redis 中对应的记录
func RemoveCachedFileAndRedisKey(cid string) {
	redisClient := redis.GetClient()
	redisClient.Del(FileKeyPrefix + cid)
	redisClient.ZRem(FileCacheSetKey, cid)

	_ = os.Remove(GetCacheFilePath(cid))
}

// 下载文件到本地缓存，并添加记录到 Redis 中
func DownloadFileToCache(cid string) (err error) {
	rpcClient, _ := utils.GetClientInstance()
	lg := utils.GetLogger()

	stream, meta, err := rpcClient.StreamRead(cid)
	if err != nil {
		lg.Errorf("An error occurred while downloading %s, %v", cid, err)
		return
	}
	defer stream.Close()

	dst, err := os.Create(GetCacheFilePath(cid))
	if err != nil {
		return
	}
	defer dst.Close()

	// 保存文件，并在保存成功后，写入文件信息到缓存
	_, err = io.Copy(dst, stream)
	if err == nil {
		c := CachedFile{
			Hash:     cid,
			Name:     meta.FName,
			Size:     meta.FSize,
			MimeType: mime.TypeByExtension(filepath.Ext(meta.FName)),
		}
		AddCachedFileToRedis(cid, c)
	}
	return
}
