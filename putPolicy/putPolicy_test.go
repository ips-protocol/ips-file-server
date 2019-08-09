package putPolicy

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

/**
 * 测试编码生成上传 Token
 */
func TestEncodePutPolicy(t *testing.T) {
	appClient := AppClient{
		AccessKey: "lfyMRgbefeeFPxbwAgFJyKaNXLQtURnv",
		SecretKey: "eZZuoTFPkMOebV0mlQxzrjsuUBqHcoV8WjNV2ejXgtN72myc",
	}

	// 1. Put policy content
	policy := PutPolicy{
		Deadline: 1563295798,
		//CallbackUrl:         "http://ipweb.io/app/file/upload-callback",
		//CallbackUrl:         "http://localhost:8081/",
		//CallbackBody:        "name=$(fname)&size=$(fsize)&hash=$(hash)&width=$(width)&height=$(height)&duration=$(duration)",
		//PersistentNotifyUrl: "http://localhost:8081/",
		//PersistentOps:       "convertVideo,videoThumb",
		//ReturnBody:          `{"name": "$(fname)", "size": $(fsize), "hash": "$(hash)"}`,
		//ReturnUrl:           "http://localhost:8081",
		ClientKey: "22ad47ccff00d3e672d8b4f3d7b2ff695805d4dce1f4a8f5d02780304f1a4862",
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
