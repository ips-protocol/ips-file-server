package fileCache

import (
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/utils"
	"os"
	"path/filepath"
)

// 获取临时目录（ ./tmp/cache/ ）下所有文件的总大小
func CalcTotalCacheSize() int64 {
	cacheDir := utils.GetCacheDir()
	var totalSize int64 = 0

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0
	}

	return totalSize
}

// 删除旧的缓存文件，减少缓存文件所占用的体积，以达到节约磁盘空间的目的
// 返回最终被删除的缓存文件数量
func ReduceCaches(reduceToFitSize int64) int {
	var filesHaveBeenRemoved int = 0
	for {
		totalSize := CalcTotalCacheSize()
		if totalSize <= reduceToFitSize {
			break
		}

		// 每次获取 5 个最旧的文件，并将其删除
		fileHashes, err := redisdb.GetClient().ZRange(FileCacheSetKey, 0, 5).Result()
		if err != nil {
			return filesHaveBeenRemoved
		}

		for _, cid := range fileHashes {
			RemoveCachedFileAndRedisKey(cid)
			filesHaveBeenRemoved++
		}
	}

	return filesHaveBeenRemoved
}
