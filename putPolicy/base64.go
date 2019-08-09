package putPolicy

import "encoding/base64"

//
// Safe base 64 encode
//
// This will use base64 to encode the src string,
// and replace "+" to "-", and "/" to "_"
//
func UrlSafeBase64Encode(src string) string {
	bytes := []byte(src)
	return base64.URLEncoding.EncodeToString(bytes)
}

func UrlSafeBase64Decode(src string) string {
	ret, err := base64.URLEncoding.DecodeString(src)
	if err != nil {
		return string(ret)
	}

	return string(ret)
}
