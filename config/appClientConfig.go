package config

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
)

// 第三方应用密钥对象
type AppClient struct {
	AccessKey   string `json:"access_key" yaml:"access_key"`
	SecretKey   string `json:"secret_key" yaml:"secret_key"`
	Description string `json:"description,omitempty"`
}

// 以 Map 的形式保存 AppClients 中 AccessKey 与 AppClient 的对应关系，方便查找
var appClientsMap map[string]AppClient

//
// 根据 AccessKey 获取 AppClient 对象
//
func GetClientByAccessKey(accessKey string) (AppClient, error) {
	if appClientsMap == nil {
		// 遍历所有 Clients，并添加到 Map
		appClientsMap = make(map[string]AppClient)
		for _, c := range GetConfig().Clients {
			appClientsMap[c.AccessKey] = c
		}
	}

	client, ok := appClientsMap[accessKey]
	if ok == false {
		return AppClient{}, errors.New("client does not exist")
	}

	return client, nil
}

//
// 签名策略字符串
//
func (client *AppClient) SignContent(encodedPutPolicy string) string {
	h := hmac.New(sha1.New, []byte(client.SecretKey))
	h.Write([]byte(encodedPutPolicy))

	return hex.EncodeToString(h.Sum(nil))
}
