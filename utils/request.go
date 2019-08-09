package utils

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// 表单类型
const RequestContentTypeFormUrlencoded = "application/x-www-form-urlencoded"

// JSON 类型
const RequestContentTypeJson = "application/json"

// 发送 Post 请求，并返回响应主体。服务端返回非 2xx 响应时抛出错误
func RequestPost(url string, body string, contentType string) (responseBody string, err error) {
	client := &http.Client{
		Timeout: time.Second * 30, // 默认请求超时时间为 30 秒
	}

	lg := GetLogger()

	lg.Debugf("Send request. Url: %s, method: POST, contentType: %s", url, contentType)
	lg.Debugf("Request body: ", body)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("User-Agent", "IPWeb SDK")

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	responseBody = string(respBody)

	lg.Debugf("Server response %d, %s", resp.StatusCode, responseBody)

	// 解析响应的状态码，如果状态码不在 200 到 299 之间，则返回一个错误
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		err = errors.New(resp.Status)
	}

	return
}
