package persistent

import (
	_redis "github.com/go-redis/redis"
	"github.com/ipweb-group/file-server/db/redis"
	"time"
)

// 待转换的任务队列
const KeyUnprocessedQueue = "IPWEB:videoConverter:unprocessed"

// 转换中的任务的哈希表
const KeyProcessingMap = "IPWEB:videoConverter:processing"

// 转换失败的任务的哈希表，值为失败时的控制台输出
const KeyFailedMap = "IPWEB:videoConverter:failed"

// 任务详情的 Key 前缀
const KeyTaskPrefix = "IPWEB:task:"

// 写入视频任务到 Redis 队列
func AddTaskToUnprocessedQueue(task *Task) {
	redisClient := redis.GetClient()
	redisClient.RPush(KeyUnprocessedQueue, task.Cid)

	// 写入视频任务详情
	redisClient.Set(KeyTaskPrefix+task.Cid, task.ToJSON(), 0)
}

// 获取第一个未处理的任务
func GetFirstUnprocessedTask() *Task {
	redisClient := redis.GetClient()
	// 获取第一个未处理任务的 CID
	cid, err := redisClient.LPop(KeyUnprocessedQueue).Result()
	if err != nil {
		return nil
	}

	// 根据 CID 查找详情记录
	key := KeyTaskPrefix + cid
	val, err := redisClient.Get(key).Result()
	if err != _redis.Nil && err != nil {
		return nil
	}

	return UnmarshalTask(val)
}

// 把任务添加到处理中的 Hash 表中
func AddTaskToProcessingMap(task *Task) {
	redisClient := redis.GetClient()
	redisClient.HSet(KeyProcessingMap, task.Cid, time.Now().Unix())
}

// 把失败的任务添加到失败的 Hash 表中，并保存失败时的控制台输出
func AddFailedTask(task *Task, stdoutContent string) {
	redisClient := redis.GetClient()
	redisClient.HSet(KeyFailedMap, task.Cid, stdoutContent)
}

// 移除正在执行的任务
func RemoveProcessingTask(task *Task) {
	redis.GetClient().HDel(KeyProcessingMap, task.Cid)
}

// 移除任务
func RemoveTask(task *Task) {
	redis.GetClient().Del(KeyTaskPrefix + task.Cid)
}
