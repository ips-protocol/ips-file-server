package ossClient

import "github.com/aliyun/aliyun-oss-go-sdk/oss"

// 重写 oss.GetObject 方法，用于返回 response 类型的数据
func GetObjectToResponse(bucket *oss.Bucket, objectKey string, options ...oss.Option) (*oss.Response, error) {
	result, err := bucket.DoGetObject(&oss.GetObjectRequest{ObjectKey: objectKey}, options)
	if err != nil {
		return nil, err
	}

	return result.Response, nil
}
