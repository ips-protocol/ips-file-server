package putPolicy

import (
	"fmt"
	"github.com/ipweb-group/file-server/config"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

/**
 * 测试编码生成上传 Token
 */
func TestEncodePutPolicy(t *testing.T) {
	appClient := config.AppClient{
		AccessKey: "lfyMRgbefeeFPxbwAgFJyKaNXLQtURnv",
		SecretKey: "eZZuoTFPkMOebV0mlQxzrjsuUBqHcoV8WjNV2ejXgtN72myc",
	}

	// 1. Put policy content
	policy := PutPolicy{
		Deadline: int32(time.Now().Unix()) + (86400 * 365 * 10),
		EndUser:  "AVATAR",
		//CallbackUrl:         "https://ipweb.io/app/file/upload-callback?auth_token=97be2dd2637c4ba6b3eb46cc83a0bfdb",
		//CallbackUrl:         "http://localhost:8081/",
		//CallbackBody:        "fname=$(fname)&fsize=$(fsize)&mimeType=$(mimeType)&endUser=$(endUser)&hash=$(hash)&width=$(width)&height=$(height)&duration=$(duration)&title=%E9%87%91%E9%B8%A1%E6%B9%96%E9%9F%B3%E4%B9%90%E5%96%B7%E6%B3%89&description=&category=art",
		//PersistentNotifyUrl: "https://ipweb.io/app/file/persistent-callback?auth_token=97be2dd2637c4ba6b3eb46cc83a0bfdb",
		//PersistentOps:       "convertVideo",
		ReturnBody: `{"hash": "$(hash)", "size": $(fsize), "type": "$(mimeType)"}`,
		//ReturnUrl:           "http://localhost:8081",
		//ClientKey: "22ad47ccff00d3e672d8b4f3d7b2ff695805d4dce1f4a8f5d02780304f1a4862",
		FSizeLimit: 20 << 20,
	}

	result := policy.Make(appClient)

	fmt.Println(result)

	//assert.Equal(t, "lfyMRgbefeeFPxbwAgFJyKaNXLQtURnv:179ee567b96a98424ef3decfa3ac328dc7fc3b49:eyJkZWFkbGluZSI6MTU2MzI5NTc5OCwiY2FsbGJhY2tVcmwiOiJodHRwOi8vbG9jYWxob3N0OjgwODEiLCJjYWxsYmFja0JvZHkiOiJuYW1lPSQoZm5hbWUpXHUwMDI2c2l6ZT0kKGZzaXplKVx1MDAyNmhhc2g9JChoYXNoKVx1MDAyNndpZHRoPSQoaW1hZ2VXaWR0aClcdTAwMjZoZWlnaHQ9JChpbWFnZUhlaWdodCkiLCJwZXJzaXN0ZW50T3BzIjoiY29udmVydFZpZGVvIiwicGVyc2lzdGVudE5vdGlmeVVybCI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MSJ9", result)
}

/**
 * 测试解码上传 Token
 */
func TestDecodePutPolicy(t *testing.T) {
	policyString := "lfyMRgbefeeFPxbwAgFJyKaNXLQtURnv:67c30737ddd80ee013ae5321dfdeba51e0475058:eyJkZWFkbGluZSI6MTU2NDkyOTQ5NiwiZW5kVXNlciI6IkIxZ2MzcjJRNyIsImNhbGxiYWNrVXJsIjoiaHR0cHM6Ly9pcHdlYi5pby9hcHAvZmlsZS91cGxvYWQtY2FsbGJhY2s_YXV0aF90b2tlbj0yZTcxNzhjZDU3Y2I0MjVjODAwZjJlZGRlZGViYzcwMSIsImNhbGxiYWNrQm9keSI6ImZuYW1lPSQoZm5hbWUpJmZzaXplPSQoZnNpemUpJm1pbWVUeXBlPSQobWltZVR5cGUpJmVuZFVzZXI9JChlbmRVc2VyKSZoYXNoPSQoaGFzaCkmd2lkdGg9JCh3aWR0aCkmaGVpZ2h0PSQoaGVpZ2h0KSZkdXJhdGlvbj0kKGR1cmF0aW9uKSZ0aXRsZT1TZWNoJTIwLSUyME90cm8lMjBUcmFnbyUyMGZ0LiUyMERhcmVsbCUyMChWaWRlbyUyME9maWNpYWwpJmRlc2NyaXB0aW9uPVZpZGVvIiwiZlNpemVMaW1pdCI6MjA5NzE1MjAwLCJwZXJzaXN0ZW50T3BzIjoiY29udmVydFZpZGVvIiwicGVyc2lzdGVudE5vdGlmeVVybCI6Imh0dHBzOi8vaXB3ZWIuaW8vYXBwL2ZpbGUvcGVyc2lzdGVudC1jYWxsYmFjaz9hdXRoX3Rva2VuPTJlNzE3OGNkNTdjYjQyNWM4MDBmMmVkZGVkZWJjNzAxIn0="
	decodedPutPolicy, err := DecodePutPolicyString(policyString)

	assert.NoError(t, err)
	assert.Equal(t, decodedPutPolicy.PutPolicy.CallbackUrl, "http://localhost:8081")
	assert.Equal(t, decodedPutPolicy.PutPolicy.CallbackBody, "name=$(fname)&size=$(fsize)&hash=$(hash)&width=$(imageWidth)&height=$(imageHeight)")
}

// 生成用于 IPWEB 中头像上传使用的上传策略
func TestEncodePutPolicyForAvatarUpload(t *testing.T) {
	appClient := config.AppClient{
		AccessKey: "lfyMRgbefeeFPxbwAgFJyKaNXLQtURnv",
		SecretKey: "eZZuoTFPkMOebV0mlQxzrjsuUBqHcoV8WjNV2ejXgtN72myc",
	}

	policy := PutPolicy{
		Deadline:   int32(time.Now().Unix()) + (86400 * 365 * 10),
		EndUser:    "AVATAR",
		ReturnBody: `{"hash": "$(hash)", "size": $(fsize), "type": "$(mimeType)"}`,
		FSizeLimit: 20 << 20,
	}

	result := policy.Make(appClient)

	fmt.Println(result)
	// lfyMRgbefeeFPxbwAgFJyKaNXLQtURnv:ddb2e89a38845c9e34f8f1e192cb06fe25dd301f:eyJkZWFkbGluZSI6MTg5MTMyOTUyMywicmV0dXJuQm9keSI6IntcImhhc2hcIjogXCIkKGhhc2gpXCIsIFwic2l6ZVwiOiAkKGZzaXplKSwgXCJ0eXBlXCI6IFwiJChtaW1lVHlwZSlcIn0iLCJlbmRVc2VyIjoiQVZBVEFSIiwiZlNpemVMaW1pdCI6MjA5NzE1MjB9
}
