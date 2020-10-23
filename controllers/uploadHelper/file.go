package uploadHelper

import (
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/utils"
	"github.com/ipweb-group/go-sdk/rpc"
	"io"
	"io/ioutil"
	"mime/multipart"
)

func GetFileHash(clientKey string, file io.Reader) (hash string, err error) {
	// 未传入 key 的时候从配置文件中取默认的 key
	if clientKey == "" {
		clientKey = config.GetConfig().NodeConf.ContractConf.ClientKeyHex
	}

	hash, err = rpc.GetCidByClientKey(clientKey, file)

	return
}

// 写入上传文件到临时文件，并返回临时文件的绝对路径
func WriteTmpFile(file multipart.File, ext string) (path string, err error) {
	tmpDir := utils.GetTmpDir()

	_, err = file.Seek(0, 0)
	if err != nil {
		return
	}

	// 把上传内容写入到临时文件中
	f, err := ioutil.TempFile(tmpDir, "uploading-*"+ext)
	if err != nil {
		return
	}
	_, err = io.Copy(f, file)
	if err != nil {
		return
	}
	_ = f.Close()

	path = f.Name()
	return
}
