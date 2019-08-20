package fileRecord

import (
	"context"
	"github.com/google/uuid"
	"github.com/ipweb-group/file-server/externals/mongodb"
	"github.com/ipweb-group/file-server/putPolicy"
	"github.com/ipweb-group/file-server/putPolicy/mediaHandler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

type FileRecord struct {
	Id                  string                 `bson:"_id"`
	Hash                string                 `bson:"hash"`
	Client              string                 `bson:"client"`
	Filename            string                 `bson:"filename"`
	MimeType            string                 `bson:"mime_type"`
	Size                int64                  `bson:"size"`
	PutPolicy           putPolicy.PutPolicy    `bson:"put_policy"`
	MediaInfo           mediaHandler.MediaInfo `bson:"media_info"`
	VideoConvertJobInfo VideoConvertJobInfo    `bson:"video_convert_job_info"`
	CreatedAt           primitive.DateTime     `bson:"created_at"`
}

// 视频转码状态及任务信息
type VideoConvertJobInfo struct {
	JobId       string             `json:"jobId" bson:"job_id"`
	RequestId   string             `json:"requestId" bson:"request_id"`
	State       string             `json:"state" bson:"state"`
	CompletedAt primitive.DateTime `json:"-" bson:"completed_at"`
}

func GetCollection() *mongo.Collection {
	return mongodb.GetDB().Collection("file_records")
}

func (receiver *FileRecord) Insert() (id string, err error) {
	randomUUID, _ := uuid.NewRandom()
	receiver.Id = randomUUID.String()
	receiver.CreatedAt = primitive.DateTime(time.Now().Unix() * 1000)

	_, err = GetCollection().InsertOne(context.Background(), receiver)
	if err != nil {
		return
	}

	id = receiver.Id
	return
}

// 更新视频转码的任务 ID （用于 OSS 上传完成，并发送转码服务后）
func UpdateVideoJobId(fileRecordId, videoJobId string) error {
	filter := bson.M{"_id": fileRecordId}
	update := bson.M{
		"$set": bson.M{"video_convert_job_info.job_id": videoJobId},
	}
	_, err := GetCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
}
