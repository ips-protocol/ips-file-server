package controllers

import (
	"github.com/ipweb-group/file-server/utils"
	"github.com/kataras/iris"
)

type ListController struct {
}

func (receiver ListController) GetList(ctx iris.Context) {
	client, _ := utils.GetClientInstance()
	_, _ = ctx.JSON(client.Nodes)
}
