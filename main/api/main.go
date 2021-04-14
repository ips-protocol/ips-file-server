package main

import (
	"context"
	"github.com/ipsfile-server/config"
	"github.com/ips/file-server/controllers"
	"github.com/ips/file-server/externals/mongodb"
	"github.com/ips/file-server/externals/ossClient"
	"github.com/ips/file-server/externals/redisdb"
	"github.com/ips/file-server/utils"
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
	logger := app.Logger()
	logger.SetTimeFormat("2006/01/02 15:04:05")
	utils.SetLogger(app.Logger())

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

		logger.Info("Server is stopping...")

		// 关闭 Redis 连接
		if err := redisdb.Close(); err != nil {
			logger.Warnf("Close redis connection failed, %v", err)
		}

		// 关闭 Mongo 连接
		if err := mongodb.Close(); err != nil {
			logger.Warn("Close mongodb connection failed, ", err.Error())
		}

		_ = app.Shutdown(ctx)
	})

	_ = app.Run(iris.Addr(config.GetConfig().Server.HttpHost), iris.WithoutInterruptHandler)
}

// 构建路由
func routers(app *iris.Application) {
	// Version 1
	v1 := app.Party("/v1")
	{
		uploadController := controllers.UploadController{}
		v1.Post("/upload", uploadController.Upload)
		v1.Options("/upload", uploadController.Upload)

		downloadController := controllers.DownloadController{}
		v1.Get("/file/{cid:string}", downloadController.StreamedDownload)
		v1.Get("/file/{cid:string}/{operation:string}", downloadController.StreamedDownload)

		deleteController := controllers.DeleteController{}
		v1.Post("/file/delete", deleteController.Delete)

		//listController := controllers.ListController{}
		//v1.Get("/nodes", listController.GetList)

		v1.Post("/mts/subscriber", controllers.MTSController{}.Subscriber)
	}
}
