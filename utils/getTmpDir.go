package utils

import (
	"log"
	"os"
	"path"
	"path/filepath"
)

// 初始化临时目录
// 自动创建 tmp 和 tmp/cache 两个临时文件夹
func InitTmpDir() {
	_dir, _ := filepath.Abs("./tmp")
	if !PathExists(_dir) {
		err := os.Mkdir(_dir, os.ModePerm)
		if err != nil {
			log.Fatalf("[ERROR] Create tmp dir failed [%v]", err)
		}
	}

	_cacheDir := path.Join(_dir, "cache")
	if !PathExists(_cacheDir) {
		err := os.Mkdir(_cacheDir, os.ModePerm)
		if err != nil {
			log.Fatalf("[ERROR] Create cache dir failed, %v", err)
		}
	}
}

// 获取临时目录
func GetTmpDir() string {
	_dir, _ := filepath.Abs("./tmp")
	return _dir
}

// 获取缓存目录
func GetCacheDir() string {
	_dir, _ := filepath.Abs("./tmp/cache")
	return _dir
}

// 判断文件夹是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}
