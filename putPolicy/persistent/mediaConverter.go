package persistent

import (
	"encoding/json"
	"github.com/ipweb-group/file-server/utils"
	"io"
	"time"
)

// 处理多媒体文件的格式转换
func ConvertMediaJob() {
	utils.GetLogger().Info("Convert media job is started")

	// 每隔一段时间检查一次队列，并在有任务时执行任务
	for {
		time.Sleep(1 * time.Second)
		go convertMedia()
	}
}

func convertMedia() {
	// 获取第一个转换任务，并添加到转换中的 Hash 表中
	task := GetFirstUnprocessedTask()
	if task == nil {
		return
	}

	lg := utils.GetLogger()
	lg.Infof("Convert task detected. Hash is %s, Ops is %s, NotifyUrl is %s", task.Cid, task.PersistentOps, task.PersistentNotifyUrl)

	// 有任务时，添加任务到正在转换的 Hash 表中
	AddTaskToProcessingMap(task)

	// 启动转换任务
	processResults := task.ProcessPersistent()

	// 移动当前文件到缓存
	AddResultFileToCache(task.Cid, task.FilePath)

	isAllConvertTaskSucceed := true
	for _, r := range processResults {
		if r.Code != CodeSuccess {
			isAllConvertTaskSucceed = false
		}
	}

	// 移除 Redis 缓存
	RemoveProcessingTask(task)
	if isAllConvertTaskSucceed {
		RemoveTask(task)
	}

	// 完成后回调（无论成功还是失败）
	go func() {
		requestBody := NotifyRequestBody{
			Hash:    task.Cid,
			Results: processResults,
		}

		stringContent, _ := json.Marshal(requestBody)
		responseBody, err := utils.RequestPost(task.PersistentNotifyUrl, string(stringContent), utils.RequestContentTypeJson)
		if err != nil {
			lg.Warnf("Callback failed in persistent process, %v", err)
		}
		lg.Debugf("Callback in persistent process responds: %s", responseBody)
	}()
}

// 关闭文件。发生错误时不做任何处理
func closeFile(file io.Closer) {
	err := file.Close()
	if err != nil {
		utils.GetLogger().Warn("An error occurred while closing file")
	}
}
