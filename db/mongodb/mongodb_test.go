package mongodb

import (
	"context"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/putPolicy"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"testing"
)

var conf = config.MongoConfig{
	ConnectionUri: "mongodb://127.0.0.1:27017",
	Db:            "ipweb-storage",
}

func init() {
	Connect(conf)
}

func TestConnect(t *testing.T) {
	Connect(conf)
}

func TestInsertFile(t *testing.T) {
	file := File{
		Id:               "testing2222",
		OriginalFilename: "test.pdf",
		MimeType:         "application/pdf",
		FileSize:         300,
		PutPolicy:        putPolicy.PutPolicy{},
		ConvertStatus:    0,
		RelatedFiles:     nil,
	}

	file.Insert()
}

func TestFindFile(t *testing.T) {
	c := GetDB().Collection("files")
	var f File
	err := c.FindOne(context.Background(), bson.M{"_id": "hello"}).Decode(&f)
	if err == mongo.ErrNoDocuments {
		log.Print("No document")
	}
	if err != nil {
		log.Fatal(err)
	}

}
