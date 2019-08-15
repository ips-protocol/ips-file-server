package redisdb

import (
	"github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/config"
	"log"
)

var redisClient *redis.Client

func Connect() {
	if redisClient == nil {
		redisConf := config.GetConfig().Redis

		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConf.Addr,
			Password: redisConf.Password,
			DB:       redisConf.DB,
			OnConnect: func(conn *redis.Conn) error {
				log.Print("[INFO] Redis connection established")
				return nil
			},
		})

		// 尝试 ping redis server，失败时表示连接 Redis 失败
		_, err := redisClient.Ping().Result()
		if err != nil {
			log.Print("[ERROR] Connect to Redis server failed")
			panic(err)
		}
	}

	return
}

func GetClient() *redis.Client {
	if redisClient == nil {
		Connect()
	}

	return redisClient
}

func Close() (err error) {
	if redisClient != nil {
		err = redisClient.Close()
	}
	return
}
