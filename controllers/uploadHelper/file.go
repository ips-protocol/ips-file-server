package uploadHelper

import (
	"github.com/ipweb-group/file-server/utils"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
)

func GetFileHash(clientKey string, file io.Reader) (hash string, err error) {
	rpcClient, _ := utils.GetClientInstance()

	if clientKey != "" {
		hash, err = rpcClient.GetCidByClientKey(clientKey, file)
	} else {
		hash, err = rpcClient.GetCid(file)
	}
	return
}

// 写入上传文件到临时文件，并返回临时文件的绝对路径
func WriteTmpFile(file multipart.File, cid string, ext string) (path string, err error) {
	tmpDir := utils.GetTmpDir()
	path = tmpDir + "/" + cid + ext

	_, err = file.Seek(0, 0)
	if err != nil {
		return
	}

	// 把上传内容写入到临时文件中
	f, err := ioutil.TempFile(tmpDir, "uploading-*")
	if err != nil {
		return
	}
	_, err = io.Copy(f, file)
	if err != nil {
		return
	}
	_ = f.Close()

	// 重命名临时文件为具体的文件名。如果指定的文件已存在，则会直接覆盖
	err = os.Rename(f.Name(), path)
	return
}
