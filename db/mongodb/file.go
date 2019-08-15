package mongodb

import (
	"context"
	"github.com/ipweb-group/file-server/putPolicy"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type File struct {
	Id               string              `bson:"_id"`
	OriginalFilename string              `bson:"original_filename"`
	MimeType         string              `bson:"mime_type"`
	FileSize         int64               `bson:"file_size"`
	PutPolicy        putPolicy.PutPolicy `bson:"put_policy"`
	MediaInfo        MediaInfo           `bson:"media_info"`
	ConvertStatus    int8                `bson:"convert_status"` // TODO 这个字段感觉好像没什么用……
	RelatedFiles     map[string]string   `bson:"related_files"`
}

type MediaInfo struct {
	Width    int    `json:"width" bson:"width"`
	Height   int    `json:"height" bson:"height"`
	Duration string `json:"duration" bson:"duration"`
	Type     string `json:"type" bson:"type"`
}

const (
	FileCovertStatusNone       = 0
	FileCovertStatusProcessing = 1
	FileCovertStatusDone       = 2
)

func (f *File) Insert() (err error) {
	_, err = GetFileCollection().InsertOne(context.TODO(), f)
	return
}

// 获取文件的 mongo collection
func GetFileCollection() *mongo.Collection {
	return GetDB().Collection("files")
}

// 根据文件 hash 获取文件对象
func GetFileByHash(hash string) (f File, err error) {
	err = GetFileCollection().FindOne(context.Background(), bson.M{"_id": hash}).Decode(&f)
	return
}
