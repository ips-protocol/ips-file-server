package aliyun

import (
	"fmt"
	"github.com/aliyun/aliyun-mns-go-sdk"
	"testing"
)

type appConf struct {
	Url             string `json:"url"`
	AccessKeyId     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
}

// MNS 创建 HTTPS 类型的订阅（网站后台不支持创建 HTTPS 类型，仅能通过代码创建）
func TestCreateSubscriber(t *testing.T) {
	conf := appConf{
		Url:             "http://1482808951674974.mns.cn-hongkong.aliyuncs.com/",
		AccessKeyId:     "LTAIyM1HpVjLYkMh",
		AccessKeySecret: "U90pZ80lt8k0iF70phEOpeZRuIZZKh",
	}

	client := ali_mns.NewAliMNSClient(conf.Url,
		conf.AccessKeyId,
		conf.AccessKeySecret)

	// 3. subscribe to topic, the endpoint is set to be a queue in this sample
	topic := ali_mns.NewMNSTopic("ipweb-mts-notify", client)
	sub := ali_mns.MessageSubsribeRequest{
		Endpoint:            "https://up.ipweb.io/v1/mts/subscriber",
		NotifyContentFormat: ali_mns.XML,
		NotifyStrategy:      ali_mns.BACKOFF_RETRY,
	}

	// topic.Unsubscribe("SubscriptionNameA")
	err := topic.Subscribe("ipweb-mts-subscriber", sub)
	if err != nil && !ali_mns.ERR_MNS_SUBSCRIPTION_ALREADY_EXIST_AND_HAVE_SAME_ATTR.IsEqual(err) {
		fmt.Println(err)
		return
	}

}
