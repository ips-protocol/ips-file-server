package mongodb

import (
	"context"
	"github.com/ipweb-group/file-server/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var mongoClient *mongo.Client
var db *mongo.Database

func Connect(conf config.MongoConfig) {
	// Set client options
	clientOptions := options.Client().ApplyURI(conf.ConnectionUri).SetConnectTimeout(10 * time.Second)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal("[ERROR] Connect to mongodb failed, ", err.Error())
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal("[ERROR] Connect to mongodb failed, ", err.Error())
	}

	mongoClient = client

	log.Print("[INFO] Mongodb connection established")

	db = client.Database(conf.Db)
}

func GetClient() *mongo.Client {
	return mongoClient
}

func GetDB() *mongo.Database {
	if db == nil {
		log.Fatal("DB is nil")
	}
	return db
}

func Close() (err error) {
	if mongoClient != nil {
		err = mongoClient.Disconnect(context.TODO())
	}
	return
}
