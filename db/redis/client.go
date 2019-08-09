package redis

import (
	"github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/config"
	"github.com/ipweb-group/file-server/utils"
)

var redisClient *redis.Client

func GetClient() *redis.Client {
	if redisClient == nil {
		redisConf := config.GetConfig().Redis

		redisClient = redis.NewClient(&redis.Options{
			Addr:     redisConf.Addr,
			Password: redisConf.Password,
			DB:       redisConf.DB,
			OnConnect: func(conn *redis.Conn) error {
				utils.GetLogger().Info("Redis connection established")
				return nil
			},
		})

		// 尝试 ping redis server，失败时表示连接 Redis 失败
		_, err := redisClient.Ping().Result()
		if err != nil {
			utils.GetLogger().Error("Connect to Redis server failed")
			panic(err)
		}
	}

	return redisClient
}
