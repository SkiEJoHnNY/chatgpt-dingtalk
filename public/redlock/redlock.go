package redlock

import (
	"context"
	"github.com/SkiEJoHnNY/chatgpt-dingtalk/config"
	"github.com/redis/go-redis/v9"

	"github.com/bsm/redislock"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

var client *redis.Client
var locker *redislock.Client
var once sync.Once

const RetryTime int = 3

var BackoffTime time.Duration = config.LoadConfig().ContextGapTime/time.Duration(RetryTime) + 1
var options = &redislock.Options{
	RetryStrategy: redislock.LimitRetry(redislock.LinearBackoff(BackoffTime), RetryTime),
	Metadata:      "",
}

const redisAddr = "redis-chatgpt-dingtalk"

func init() {
	once.Do(func() {
		client = redis.NewClient(&redis.Options{
			Network:  "tcp",
			Addr:     redisAddr + ":6379",
			Password: "whyaliyun",
		})
		log.Debug(client)
		locker = redislock.New(client)
	})
}

func Get(key string, expiration time.Duration, xOptions *redislock.Options) (*redislock.Lock, error) {
	if xOptions == nil {
		xOptions = options
	}
	lock, err := locker.Obtain(context.Background(), key, expiration, xOptions)
	if err == redislock.ErrNotObtained {
		log.Debug(key, "cant obtain redislock")
	} else if err != nil {
		log.Fatal(err)
	}
	return lock, err
}
