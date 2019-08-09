package persistent

import (
	"encoding/json"
	"github.com/ipweb-group/file-server/putPolicy/mediaHandler"
	"github.com/ipweb-group/file-server/utils"
	"regexp"
	"strings"
)

// 持久化任务处理器
type Task struct {
	Cid                 string                 `json:"cid"`
	FilePath            string                 `json:"filePath"` // 临时文件保存的路径
	PersistentOps       string                 `json:"persistentOps"`
	PersistentNotifyUrl string                 `json:"persistentNotifyUrl"`
	MediaInfo           mediaHandler.MediaInfo `json:"mediaInfo"`
	ClientKey           string                 `json:"clientKey"` // 上传者的密钥
}

const (
	// 视频转码任务，将视频转换为 h264/AAC 格式以供移动端播放使用
	TaskConvertVideo = "convertVideo"
	// 获取视频第一帧缩略图
	TaskVideoThumb = "videoThumb"
)

// 解析 JSON 字符串为 Task 类型的数据
func UnmarshalTask(str string) *Task {
	ret := &Task{}

	err := json.Unmarshal([]byte(str), ret)
	if err != nil {
		return nil
	}

	return ret
}

// 检查任务是否需要被添加到转换队列中，如果不需要被转换，则会返回允许删除临时文件的标识
func (h *Task) CheckShouldQueueTask() bool {
	// 目前所支持的转换全都是针对视频的，所以只要判断文件为视频类型，并且 PersistentOps 不为空，
	// 就将转换任务添加到队列中
	if h.PersistentOps != "" {
		if match, _ := regexp.MatchString("video/.*", h.MediaInfo.MimeType); match {
			return false
		}
	}

	return true
}

// 添加任务添加到未处理队列中
func (h *Task) Queue() {
	AddTaskToUnprocessedQueue(h)
	utils.GetLogger().Infof("Video file is queued to redis: %s", h.FilePath)
}

// 将任务转换为 JSON 字符串（用于保存到 Redis）
func (h *Task) ToJSON() string {
	j, _ := json.Marshal(h)
	return string(j)
}

// 获取解析后的每个持久化任务名称，返回单个名称组成的数组
func (h *Task) GetPersistentOps() []string {
	return strings.Split(h.PersistentOps, ",")
}

// 执行转换任务
func (h *Task) ProcessPersistent() (results []Result) {
	for _, op := range h.GetPersistentOps() {
		var result Result

		switch op {
		// 转换视频任务
		case TaskConvertVideo:
			result = FfmpegCovertVideo(h)

		// 生成缩略图任务
		case TaskVideoThumb:
			result = FfmpegVideoThumb(h)
		}

		// 完成转换后，根据转换结果选择上传或者添加到错误
		if result.Code != CodeSuccess {
			AddFailedTask(h, result.Desc)

		} else {
			if result.DstHash == "" {
				dstCid, err := result.UploadConvertedFile(h)
				if err != nil {
					AddFailedTask(h, err.Error())
					result.Code = CodeFailed
					result.Desc = err.Error()

				} else {
					result.DstHash = dstCid

					// 上传完成后，移动生成的目标文件到 cache 目录中，并添加 Redis 缓存
					go result.AddResultFileToCache(h)
				}
			}
		}

		result.PersistentOp = op
		results = append(results, result)
	}

	return
}
