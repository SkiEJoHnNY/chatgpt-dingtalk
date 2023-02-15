package cache

import (
	"github.com/patrickmn/go-cache"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type GptCache struct {
	cache *cache.Cache
}

var gptCache *GptCache

var once sync.Once

func New(expiration time.Duration, cleanUp time.Duration) *GptCache {
	if gptCache != nil {
		log.Warning("Access Created GptCache. Bug!")
	}
	once.Do(func() {
		gptCache = &GptCache{
			cache: cache.New(expiration, cleanUp),
		}
		log.Info("Create GptCache successfully")
	})
	return gptCache
}
func (s *GptCache) Delete(key string) {
	s.cache.Delete(key)
}

func (s *GptCache) Set(key, value string, timeout time.Duration) {
	s.cache.Set(key, value, timeout)
}

func (s *GptCache) Get(key string) (interface{}, bool) {
	return s.cache.Get(key)
}
