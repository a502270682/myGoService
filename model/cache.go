package model

import (
	"github.com/patrickmn/go-cache"
	"time"
)

const (
	DefaultCacheExpireTime = time.Minute * 30
	DefaultCleanupInterval = time.Minute * 10
)

type CacheDriver struct {
	cache *cache.Cache
}

var cacheDriver *CacheDriver

func GetCacheDriver() *CacheDriver {
	if cacheDriver == nil {
		cacheDriver = &CacheDriver{
			cache: cache.New(DefaultCacheExpireTime, DefaultCleanupInterval),
		}
		return cacheDriver
	}
	return cacheDriver
}

func (c *CacheDriver) SetCacheWithKeyVal(key string, val interface{}) {
	c.cache.Set(key, val, DefaultCacheExpireTime)
}

func (c *CacheDriver) SetCacheWithKeyValWithExpireTime(key string, val interface{}, expireTime time.Duration) {
	c.cache.Set(key, val, expireTime)
}

func (c *CacheDriver) GetCacheValWithKey(key string) interface{} {
	ret, found := c.cache.Get(key)
	if !found || c == nil {
		return nil
	}
	return ret
}
