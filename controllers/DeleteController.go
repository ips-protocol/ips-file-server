package controllers

import (
	"github.com/ipweb-group/file-server/backgroundWorker"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/externals/mongodb/fileRecord"
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/iris"
	"go.mongodb.org/mongo-driver/bson"
)

type DeleteController struct{}

func (receiver DeleteController) Delete(ctx iris.Context) {
	accessKey := ctx.FormValue("accessKey")
	hash := ctx.FormValue("hash")
	endUser := ctx.FormValue("endUser")
	clientKey := ctx.FormValue("clientKey")

	lg := utils.GetLogger()

	_, err := config.GetClientByAccessKey(accessKey)
	if err != nil || accessKey == "" {
		throwError(iris.StatusUnprocessableEntity, "Invalid access key", ctx)
		return
	}

	if hash == "" {
		throwError(iris.StatusUnprocessableEntity, "Invalid file hash", ctx)
		return
	}

	if endUser == "" {
		throwError(iris.StatusUnprocessableEntity, "EndUser must be provided", ctx)
		return
	}

	// 根据文件 Hash、应用的 AccessKey 和用户 ID 三者联合查找文件记录并删除
	deleteResult, err := fileRecord.DeleteAllRecordByCondition(bson.D{
		{"hash", hash},
		{"client_app_id", accessKey},
		{"put_policy.endUser", endUser},
	})
	if err != nil {
		lg.Warn("Delete file records from DB failed, ", err)
		throwError(iris.StatusInternalServerError, "Delete files failed, "+err.Error(), ctx)
		return
	}
	lg.Infof("%d file records has been deleted", deleteResult.DeletedCount)

	if deleteResult.DeletedCount == 0 {
		throwError(iris.StatusNotFound, "File not found", ctx)
		return
	}

	// 删除操作已经完成，可以直接返回成功标识到客户端
	ctx.StatusCode(iris.StatusOK)
	_, _ = ctx.JSON(iris.Map{
		"success": true,
	})

	// 查找该 hash 在数据库中是否仍然存在，如果该 Hash 仍然存在，将不会彻底删除文件
	if fileRecord.HashExists(hash) {
		lg.Info("Hash already exists in database, will NOT remove file permanently")
		return
	}

	// 添加删除任务到 Queue
	task := backgroundWorker.DeleteTask{
		Hash:      hash,
		ClientKey: clientKey,
	}
	task.Enqueue()
}
