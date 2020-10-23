package uploaders

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/ipweb-group/file-server/externals/aliyun"
	"github.com/ipweb-group/file-server/externals/mongodb/fileRecord"
	"github.com/ipweb-group/file-server/externals/ossClient"
	"github.com/ipweb-group/file-server/jobs"
	"github.com/ipweb-group/file-server/utils"
	"math/rand"
	"mime"
	"path"
	"regexp"
	"strconv"
	"time"
)

type CDNUploader struct {
	Job *jobs.UploadJob
}

func (up *CDNUploader) Upload() (err error) {
	lg := utils.GetLogger()

	mimeType := mime.TypeByExtension(path.Ext(up.Job.CacheFilePath))
	bucket := ossClient.GetBucket()

	ossFilePath := "files/" + up.Job.Hash
	err = bucket.PutObjectFromFile(ossFilePath, up.Job.CacheFilePath, oss.ContentType(mimeType))
	if err != nil {
		return
	}

	lg.Info("Upload to OSS completed")

	// 如果文件是视频类型，同时启动转码服务
	if match, _ := regexp.MatchString("video/.*", mimeType); match {
		lg.Info("File is of type video, will request converting")
		go func() {
			// 随机获取视频截图的时间点，并启动截图
			randomSnapshotTime := up.calcVideoSnapshotTime()
			if err := aliyun.VideoSnapShot(ossFilePath, "converted/"+up.Job.Hash+"/snapshot.jpg", randomSnapshotTime); err != nil {
				lg.Error("[MTS] Create video snapshot failed, ", err)
			}
		}()
		go func() {
			jobId, err := aliyun.VideoCovert(ossFilePath, "converted/"+up.Job.Hash+"/playable.mp4")
			if err != nil {
				lg.Error("[MTS] Create video convert job failed, ", err)
			} else {
				// 更新 jobId 到文件记录
				if err = fileRecord.UpdateVideoJobId(up.Job.FileRecordId, jobId); err != nil {
					lg.Error("[MTS] Update convert video job id to DB failed, ", err)
				}
			}
		}()
	}
	return
}

// 计算视频截图的时间点
// 根据视频的总时长计算，取总时长位置 20% 到 50% 之间的一个随机时间点
// 未能获取时间时，默认返回 1000（毫秒）
func (up *CDNUploader) calcVideoSnapshotTime() int32 {
	var defaultRet int32 = 1000
	duration := up.Job.MediaInfo.Duration
	if duration == "" {
		return defaultRet
	}

	durationFloat64, err := strconv.ParseFloat(duration, 64)
	if err != nil {
		return defaultRet
	}

	randStart := int32(durationFloat64 * 1000 / 5)
	randEnd := int32(durationFloat64 * 1000 / 2)

	if randEnd <= randStart {
		return defaultRet
	}

	rand.Seed(time.Now().Unix())
	timeDiff := rand.Int31n(randEnd - randStart)

	return randStart + timeDiff
}
