package putPolicy

import (
	"encoding/json"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/utils"
)

type PutPolicy struct {
	Deadline            int32  `json:"deadline" bson:"deadline"`
	ReturnUrl           string `json:"returnUrl,omitempty" bson:"returnUrl,omitempty"`
	ReturnBody          string `json:"returnBody,omitempty" bson:"returnBody,omitempty"`
	EndUser             string `json:"endUser,omitempty" bson:"endUser,omitempty"`
	ClientKey           string `json:"clientKey,omitempty" bson:"clientKey,omitempty"`
	CallbackUrl         string `json:"callbackUrl,omitempty" bson:"callbackUrl,omitempty"`
	CallbackBody        string `json:"callbackBody,omitempty" bson:"callbackBody,omitempty"`
	FSizeLimit          int32  `json:"fSizeLimit,omitempty" bson:"fSizeLimit,omitempty"`
	PersistentOps       string `json:"persistentOps,omitempty" bson:"persistentOps,omitempty"`
	PersistentNotifyUrl string `json:"persistentNotifyUrl,omitempty" bson:"persistentNotifyUrl,omitempty"`
}

//
// 转换策略为 JSON 字符串
//
func (p *PutPolicy) ToJSON() (string, error) {
	str, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(str), nil
}

// 执行回调并返回回调响应内容
func (p *PutPolicy) ExecCallback(variable MagicVariable, escapeMethod string) (responseBody string, err error) {
	callbackBody := variable.ApplyMagicVariables(p.CallbackBody, escapeMethod)

	responseBody, err = utils.RequestPost(p.CallbackUrl, callbackBody, utils.RequestContentTypeFormUrlencoded)
	return
}

// 使用 PutPolicy 对象生成上传策略字符串
func (p *PutPolicy) Make(appClient config.AppClient) string {
	jsonString, err := p.ToJSON()
	if err != nil {
		return ""
	}

	encodedPutPolicy := UrlSafeBase64Encode(jsonString)
	sign := appClient.SignContent(encodedPutPolicy)

	return appClient.AccessKey + ":" + sign + ":" + encodedPutPolicy
}
