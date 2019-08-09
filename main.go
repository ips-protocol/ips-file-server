package main

import (
	"context"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/controllers"
	"github.com/ipweb-group/file-server/db/redis"
	"github.com/ipweb-group/file-server/putPolicy/persistent"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/iris"
	irisContext "github.com/kataras/iris/context"
	"time"
)

// 最大允许上传的文件大小：500MB
const MaxFileSize int64 = 500 << 20

func init() {
	// 初始化临时目录
	utils.InitTmpDir()

	// 加载配置文件
	config.LoadConfig("./config.yml")
}

func main() {
	// 初始化 Web 服务器
	app := iris.Default()
	// 设置日志实例
	utils.SetLogger(app.Logger())

	// 初始化 RPC 客户端
	rpcClient, err := utils.GetClientInstance()
	if err != nil {
		panic(err)
	}

	// 启动转换器线程
	go persistent.ConvertMediaJob()

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

		// 关闭 RPC 客户端
		err := rpcClient.Close()
		if err != nil {
			utils.GetLogger().Warnf("Close RPC Client failed, %v", err)
		}

		// 关闭 Redis 连接
		err = redis.GetClient().Close()
		if err != nil {
			utils.GetLogger().Warnf("Close redis connection failed, %v", err)
		}

		_ = app.Shutdown(ctx)
	})

	_ = app.Run(iris.Addr(config.GetConfig().Server.HttpHost), iris.WithoutInterruptHandler)

	// app.Run(iris.AutoTLS(":443", "example.com", "admin@example.com")) 可以自动配置 Lets Encrypt 证书
}

// 构建路由
func routers(app *iris.Application) {
	// Version 1
	v1 := app.Party("/v1")
	{
		uploadController := controllers.UploadController{}
		v1.Post("/upload", iris.LimitRequestBodySize(MaxFileSize), uploadController.Upload)

		downloadController := controllers.DownloadController{}
		v1.Get("/file/{cid:string}", downloadController.StreamedDownload)
	}
}