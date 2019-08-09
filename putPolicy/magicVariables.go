package putPolicy

import (
	"net/url"
	"strconv"
	"strings"
)

/**
 * 魔法变量
 */
type MagicVariable struct {
	FName    string `json:"fname"`
	FSize    int64  `json:"fsize"`
	MimeType string `json:"mimeType"`
	EndUser  string `json:"endUser"`
	Hash     string `json:"hash"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Duration string `json:"duration"`
}

const (
	EscapeJSON = "json" // js 转义
	EscapeURL  = "url"  // url encode
)

/**
 * 应用魔法变量
 *
 * 替换 returnBody 中的所有魔法变量文本并返回替换后的文本 （会对替换的值进行 URL Encode）
 */
func (variables *MagicVariable) ApplyMagicVariables(returnBody string, escape string) (ret string) {
	ret = replaceStringWithUrlEncode(returnBody, "$(fname)", variables.FName, escape)
	ret = replaceStringWithUrlEncode(ret, "$(fsize)", strconv.FormatInt(variables.FSize, 10), escape)
	ret = replaceStringWithUrlEncode(ret, "$(mimeType)", url.QueryEscape(variables.MimeType), escape)
	ret = replaceStringWithUrlEncode(ret, "$(endUser)", url.QueryEscape(variables.EndUser), escape)
	ret = replaceStringWithUrlEncode(ret, "$(hash)", variables.Hash, escape)
	ret = replaceStringWithUrlEncode(ret, "$(width)", strconv.Itoa(variables.Width), escape)
	ret = replaceStringWithUrlEncode(ret, "$(height)", strconv.Itoa(variables.Height), escape)
	ret = replaceStringWithUrlEncode(ret, "$(duration)", variables.Duration, escape)

	return
}

// 替换字符串，同时进行 url encode 转义
func replaceStringWithUrlEncode(inputString string, find string, replace string, escape string) string {
	var escapedStr string
	switch escape {
	case EscapeJSON:
		escapedStr = strings.ReplaceAll(replace, "\\", "\\\\")
		escapedStr = strings.ReplaceAll(escapedStr, "\"", "\\\"")
		break

	case EscapeURL:
		escapedStr = url.QueryEscape(replace)
		break
	}
	return strings.ReplaceAll(inputString, find, escapedStr)
}
