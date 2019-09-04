package main

import (
	"context"
	"github.com/ipweb-group/file-server/backgroundWorker"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/controllers"
	"github.com/ipweb-group/file-server/externals/mongodb"
	"github.com/ipweb-group/file-server/externals/ossClient"
	"github.com/ipweb-group/file-server/externals/redisdb"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/iris"
	irisContext "github.com/kataras/iris/context"
	"time"
)

func init() {
	// 初始化临时目录
	utils.InitTmpDir()

	// 加载配置文件
	conf := config.LoadConfig("./config.yml")

	// 连接数据库
	mongodb.Connect(conf.Mongo)
	redisdb.Connect()

	// 创建 OSS 客户端
	ossClient.GetBucket()
}

func main() {
	// 初始化 Web 服务器
	app := iris.Default()
	// 设置日志实例
	utils.SetLogger(app.Logger())

	conf := config.GetConfig()

	// 初始化 RPC 客户端
	rpcClient, err := utils.GetClientInstance()
	if err != nil {
		panic(err)
	}

	// 启动后台 Worker
	ctx, cancelBackgroundWorker := context.WithCancel(context.Background())
	go backgroundWorker.StartWorker(ctx)
	defer cancelBackgroundWorker()

	// 404 错误输出
	app.OnErrorCode(iris.StatusNotFound, func(ctx irisContext.Context) {
		_, _ = ctx.JSON(iris.Map{"error": "Document not found"})
	})

	// 构建路由
	routers(app)

	// 平滑关闭服务
	iris.RegisterOnInterrupt(func() {
		timeout := 15 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		utils.GetLogger().Info("Server is stopping...")

		// 停止后台 Worker
		cancelBackgroundWorker()

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

		_ = app.Shutdown(ctx)
	})

	if conf.Server.EnableHttps {
		_ = app.Run(iris.AutoTLS(conf.Server.HttpsHost, conf.Server.HttpsDomains, conf.Server.HttpsEmail))

	} else {
		_ = app.Run(iris.Addr(config.GetConfig().Server.HttpHost), iris.WithoutInterruptHandler)
	}
}

// 构建路由
func routers(app *iris.Application) {
	// Version 1
	v1 := app.Party("/v1")
	{
		uploadController := controllers.UploadController{}
		v1.Post("/upload", uploadController.Upload)

		downloadController := controllers.DownloadController{}
		v1.Get("/file/{cid:string}", downloadController.StreamedDownload)
		v1.Get("/file/{cid:string}/{operation:string}", downloadController.StreamedDownload)

		deleteController := controllers.DeleteController{}
		v1.Post("/file/delete", deleteController.Delete)

		listController := controllers.ListController{}
		v1.Get("/nodes", listController.GetList)

		v1.Post("/mts/subscriber", controllers.MTSController{}.Subscriber)
	}
}
