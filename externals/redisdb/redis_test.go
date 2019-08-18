package redisdb

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/config"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func init() {
	config.LoadConfig("../../config.yml")
}

func TestGetClient(t *testing.T) {
	client := GetClient()
	fmt.Println(client)

	pong, err := client.Ping().Result()
	assert.NoError(t, err)
	assert.Equal(t, pong, "PONG")
	fmt.Println(pong, err)
}

func TestExist(t *testing.T) {
	client := GetClient()
	ret, err := client.Exists("TESTING_EXIST").Result()

	assert.NoError(t, err)
	assert.Equal(t, ret, int64(0))
}

func TestZRangeByScore(t *testing.T) {
	client := GetClient()
	ret, err := client.ZRangeByScore("IPWEB:FS:QUEUE:UP", redis.ZRangeBy{
		Max:    "+Inf",
		Min:    strconv.FormatInt(time.Now().Unix(), 10),
		Offset: 0,
		Count:  1,
	}).Result()

	fmt.Println(err)
	fmt.Println(ret)
}
