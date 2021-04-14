package main

import (
	"github.com/ips/file-server/config"
	"github.com/ips/file-server/externals/mongodb"
	"github.com/ips/file-server/externals/redisdb"
	"github.com/ips/file-server/utils"
	"github.com/ips/file-server/worker"
	"github.com/kataras/golog"
)

func init() {
	// 设置日志实例
	lg := golog.New()
	lg.SetTimeFormat("2006/01/02 15:04:05")
	utils.SetLogger(lg)

	// 加载配置文件
	conf := config.LoadConfig("./config.yml")

	// 连接数据库
	mongodb.Connect(conf.Mongo)
	redisdb.Connect()
}

func main() {
	// 初始化 RPC 客户端
	rpcClient, err := utils.GetClientInstance()
	if err != nil {
		panic(err)
	}

	defer func() {
		utils.GetLogger().Info("Server is stopping...")

		// 关闭 RPC 客户端
		if err := rpcClient.Close(); err != nil {
			utils.GetLogger().Warnf("Close RPC Client failed, %v", err)
		}

		// 关闭 Redis 连接
		if err := redisdb.Close(); err != nil {
			utils.GetLogger().Warnf("Close redis connection failed, %v", err)
		}

		// 关闭 Mongo 连接
		if err := mongodb.Close(); err != nil {
			utils.GetLogger().Warn("Close mongodb connection failed, ", err.Error())
		}
	}()

	// 启动后台 Worker
	worker.StartWorker()
}
