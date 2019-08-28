package fileRecord

import (
	"context"
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
	ClientAppId         string                 `bson:"client_app_id"`
	Filename            string                 `bson:"filename"`
	MimeType            string                 `bson:"mime_type"`
	Size                int64                  `bson:"size"`
	PutPolicy           putPolicy.PutPolicy    `bson:"put_policy"`
	MediaInfo           mediaHandler.MediaInfo `bson:"media_info,omitempty"`
	VideoConvertJobInfo VideoConvertJobInfo    `bson:"video_convert_job_info,omitempty"`
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
	// 使用十六进制的 ObjectID 作为主键 ID，保存为 string 类型
	receiver.Id = primitive.NewObjectID().Hex()
	receiver.CreatedAt = primitive.DateTime(time.Now().Unix() * 1000)

	_, err = GetCollection().InsertOne(context.Background(), receiver)
	if err != nil {
		return
	}

	id = receiver.Id
	return
}

// 更新视频信息
func (receiver *FileRecord) UpdateVideoJobInfo(jobInfo VideoConvertJobInfo) error {
	receiver.VideoConvertJobInfo = jobInfo
	filter := bson.M{"_id": receiver.Id}
	update := bson.M{
		"$set": bson.M{"video_convert_job_info": jobInfo},
	}
	_, err := GetCollection().UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	return nil
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

// 根据视频转码的任务 ID 查找对应的文件记录
func GetFileRecordByVideoJobID(jobId string) (record FileRecord, err error) {
	filter := bson.M{"video_convert_job_info.job_id": jobId}
	err = GetCollection().FindOne(context.Background(), filter).Decode(&record)
	return
}

// 根据文件 Hash 获取文件记录
func GetFileRecordByHash(hash string) (record FileRecord, err error) {
	filter := bson.M{"hash": hash}
	err = GetCollection().FindOne(context.Background(), filter).Decode(&record)
	return
}

// 删除所有满足条件的记录
func DeleteAllRecordByCondition(filter interface{}) (*mongo.DeleteResult, error) {
	return GetCollection().DeleteMany(context.Background(), filter)
}

// 计算某个 Hash 的文件在数据库中的数量
func CountHash(hash string) (int64, error) {
	return GetCollection().CountDocuments(context.Background(), bson.M{"hash": hash})
}

// 判断某个 Hash 是否存在
func HashExists(hash string) bool {
	count, err := CountHash(hash)
	if err != nil {
		return false
	}
	return count > 0
}
