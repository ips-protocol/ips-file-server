package persistent

import (
	"github.com/ipweb-group/file-server/fileCache"
	"github.com/ipweb-group/file-server/utils"
	"mime"
	"os"
	"path"
	"path/filepath"
)

type NotifyRequestBody struct {
	Hash    string   `json:"hash"`    // 原始文件的 CID
	Results []Result `json:"results"` // 持久化的结果
}

type Result struct {
	Code           int    `json:"code"`         // 状态码。0 表示成功；1 表示失败
	Desc           string `json:"desc"`         // 状态对应的描述
	PersistentOp   string `json:"persistentOp"` // 持久化操作的名称
	DstHash        string `json:"dstHash"`      // 生成的目标文件的 CID
	outputFilePath string `json:"-"`            // 临时文件输出路径
}

// 成功
const CodeSuccess = 0

// 失败
const CodeFailed = 1

// 上传转换后的文件
func (r *Result) UploadConvertedFile(task *Task) (dstCid string, err error) {
	lg := utils.GetLogger()
	lg.Info("Uploading converted file to IPFS")

	file, err := os.Open(r.outputFilePath)
	if err != nil {
		lg.Errorf("Open file failed, %v", err)

	} else {
		defer closeFile(file)
		dstFileInfo, _ := file.Stat()

		// 根据是否有密钥选择对应的上传方式
		rpcClient, _ := utils.GetClientInstance()
		if task.ClientKey != "" {
			dstCid, err = rpcClient.UploadByClientKey(task.ClientKey, file, dstFileInfo.Name(), dstFileInfo.Size())
		} else {
			dstCid, err = rpcClient.Upload(file, dstFileInfo.Name(), dstFileInfo.Size())
		}

		if err != nil {
			lg.Errorf("Upload converted file failed, [%v]", err)

		} else {
			lg.Infof("Upload converted file completed, cid is %s", dstCid)
		}
	}

	return
}

// 添加转换完成后的文件到缓存中
func AddResultFileToCache(cid string, filePath string) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return
	}

	c := fileCache.CachedFile{
		Hash:     cid,
		Name:     fi.Name(),
		Size:     fi.Size(),
		MimeType: mime.TypeByExtension(filepath.Ext(fi.Name())),
	}

	err = os.Rename(filePath, path.Join(utils.GetCacheDir(), cid))
	if err != nil {
		return
	}

	fileCache.AddCachedFileToRedis(cid, c)
	utils.GetLogger().Infof("Add processed file %s to cache", cid)
}

// 添加转换完成后的文件到缓存中
func (r *Result) AddResultFileToCache(task *Task) {
	// 当目标文件的 Hash 与源文件 Hash 相同时，不作操作
	if r.DstHash == task.Cid || r.outputFilePath == "" {
		return
	}

	AddResultFileToCache(r.DstHash, r.outputFilePath)
}
