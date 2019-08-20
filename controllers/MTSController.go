package controllers

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"github.com/ipweb-group/file-server/externals/mongodb/fileRecord"
	"github.com/ipweb-group/file-server/putPolicy/persistent"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"io"
	"time"
)

type MTSController struct {
}

type notification struct {
	XMLName          xml.Name `xml:"Notification"`
	TopicOwner       string   `xml:"TopicOwner"`
	TopicName        string   `xml:"TopicName"`
	Subscriber       string   `xml:"Subscriber"`
	SubscriptionName string   `xml:"SubscriptionName"`
	MessageId        string   `xml:"MessageId"`
	MessageMD5       string   `xml:"MessageMD5"`
	Message          string   `xml:"Message"`
	PublishTime      int64    `xml:"PublishTime"`
	SigningCertURL   string   `xml:"SigningCertURL"`
}

func (mc MTSController) Subscriber(ctx iris.Context) {

	lg := ctx.Application().Logger()

	lg.Info("Get MTS callback notification")

	// 获取请求原文 （XML 格式）
	_requestBody := ctx.Request().Body
	buffer := new(bytes.Buffer)
	defer _requestBody.Close()
	_, err := io.Copy(buffer, _requestBody)
	if err != nil {
		throwError(iris.StatusUnprocessableEntity, "Cannot parse input content, "+err.Error(), ctx)
		return
	}
	body := buffer.Bytes()

	lg.Info("MTS callback notification content is:", string(body))

	// 2. 解码 XML 内容
	var xmlContent notification
	err = xml.Unmarshal(body, &xmlContent)
	if err != nil {
		throwError(iris.StatusUnprocessableEntity, "Parse input content error, "+err.Error(), ctx)
		return
	}

	// 3. 获取到 Message 的内容，并作为 JSON 格式解析
	_message := xmlContent.Message
	if _message == "" {
		throwError(iris.StatusUnprocessableEntity, "Input message cannot be empty", ctx)
		return
	}

	var jobInfo fileRecord.VideoConvertJobInfo
	err = json.Unmarshal([]byte(_message), &jobInfo)
	if err != nil {
		throwError(iris.StatusUnprocessableEntity, "Parse message failed, "+err.Error(), ctx)
		return
	}

	// 4. 查找任务 ID 对应的文件记录
	file, err := fileRecord.GetFileRecordByVideoJobID(jobInfo.JobId)
	if err != nil {
		throwError(iris.StatusNotFound, "Record not found, "+err.Error(), ctx)
		return
	}

	// 5. 更新转码状态并重新保存
	jobInfo.CompletedAt = primitive.DateTime(time.Now().Unix() * 1000)
	if err = file.UpdateVideoJobInfo(jobInfo); err != nil {
		throwError(iris.StatusInternalServerError, "Update file record failed, "+err.Error(), ctx)
		return
	}

	// 6. 检查原上传策略中是否有要求转码完成回调，如果有，则异步发送回调该请求
	if file.PutPolicy.PersistentNotifyUrl != "" {
		go func() {
			requestBody := persistent.NotifyRequestBody{
				Hash: file.Hash,
				Results: []persistent.Result{
					{
						Code:         0,
						Desc:         jobInfo.State,
						PersistentOp: "convertVideo",
					},
				},
			}

			stringContent, _ := json.Marshal(requestBody)
			responseBody, err := utils.RequestPost(file.PutPolicy.PersistentNotifyUrl, string(stringContent), utils.RequestContentTypeJson)
			if err != nil {
				lg.Warnf("Callback failed in persistent process, %v", err)
			}
			lg.Debugf("Callback in persistent process responds: %s", responseBody)
		}()
	}

	_, _ = ctx.JSON(iris.Map{
		"hash": file.Hash,
	})
}
