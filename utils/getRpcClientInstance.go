package utils

import (
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/go-sdk/rpc"
)

var clientInstance *rpc.Client

// 获取 RPC Client 的单例，此方法将在全局维护一个 client 实例，
// 避免在多个位置重复初始化 RPC Client，并简化代码结构
func GetClientInstance() (client *rpc.Client, err error) {
	if clientInstance == nil {
		clientInstance, err = rpc.NewClient(config.GetConfig().NodeConf)
		if err != nil {
			return
		}
	}

	client = clientInstance
	return
}
