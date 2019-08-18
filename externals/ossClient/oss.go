package ossClient

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/ipweb-group/file-server/config"
	"log"
)

var _client *oss.Client

func GetClient() *oss.Client {

	if _client == nil {
		ossConf := config.GetConfig().OSS

		client, err := oss.New(ossConf.Endpoint, ossConf.AccessKey, ossConf.AccessSecret)
		if err != nil {
			log.Fatal("Create OSS client failed, " + err.Error())
		}

		_client = client
	}

	return _client
}

func GetBucket() *oss.Bucket {
	client := GetClient()

	bucket, _ := client.Bucket(config.GetConfig().OSS.Bucket)
	return bucket
}
