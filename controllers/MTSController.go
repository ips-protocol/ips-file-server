package controllers

import (
	"bytes"
	"fmt"
	"github.com/kataras/iris"
	"io"
)

type MTSController struct {
}

func (mc MTSController) Subscriber(ctx iris.Context) {

	lg := ctx.Application().Logger()

	lg.Info("Get MTS callback notification")

	_requestBody := ctx.Request().Body
	buffer := new(bytes.Buffer)

	defer _requestBody.Close()

	length, err := io.Copy(buffer, _requestBody)

	lg.Info("MTS callback notification content is:")

	fmt.Println(length, err)
	fmt.Println(string(buffer.Bytes()))

}
