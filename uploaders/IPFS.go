package uploaders

import (
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/jobs"
	"github.com/ipweb-group/file-server/utils"
	"os"
	"time"
)

type IPFSUploader struct {
	Job *jobs.UploadJob
}

func (up *IPFSUploader) Upload() (err error) {
	lg := utils.GetLogger()
	rpcClient, _ := utils.GetClientInstance()
	var cid string

	lg.Infof("Prepare to upload to IPFS, cid is %s, client key is %s", up.Job.Hash, up.Job.ClientKey)
	clientKey := up.Job.ClientKey
	if clientKey == "" {
		clientKey = config.GetConfig().NodeConf.ContractConf.ClientKeyHex
	}
	cid, err = rpcClient.UploadWithPathByClientKey(clientKey, up.Job.CacheFilePath)

	if err != nil {
		lg.Warn("Upload failed, ", err.Error())
	} else {
		lg.Infof("Upload to IPFS successful. Current cid is %s, target cid is %s", up.Job.Hash, cid)
	}

	return up.afterUpload(err)
}

func (up *IPFSUploader) afterUpload(lastError error) (err error) {
	lg := utils.GetLogger()
	if lastError != nil {
		// 出错后，重试次数加一
		up.Job.RetryTimes++
		// 重试时间加一分钟，并重新保存回 Redis
		up.Job.Enqueue(time.Now().Unix() + 60)

		lg.Info("Reset failed upload job to queue, retry times set to ", up.Job.RetryTimes)
		// 重新返回旧的错误
		err = lastError

	} else {
		// 上传成功后，删除缓存文件
		redisClient := redisdb.GetClient()
		redisClient.Del(jobs.GetUploadTaskCacheKey(up.Job.FileRecordId))
		_ = os.Remove(up.Job.CacheFilePath)
		lg.Info("File cache is removed")
	}

	return
}
