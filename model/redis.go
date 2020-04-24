package model

import (
	"github.com/go-redis/redis"
	logger "github.com/sirupsen/logrus"
	"myGoService/config"
)

func (m *Model) InitRedisClient(redisDB int) {
	if m.RedisClient == nil {
		//m.RedisClient = m.connectRedisClient(redisDB)
		m.RedisClient = m.connectRedisClientWithSentinel()
	}
	if m.RedisClusterClient == nil {
		m.RedisClusterClient = m.InitRedisCluster()
	}
}

func (m *Model) connectRedisClient(redisDB int) *redis.Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.REDIS_ADDR,
		Password: "",
		DB:       redisDB,
	})
	if err := pingRedisClient(redisClient); err != nil {
		panic(err)
	}
	return redisClient
}

func (m *Model) connectRedisClientWithSentinel() *redis.Client {
	redisClient := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    config.Instance().RedisMasterName,
		SentinelAddrs: config.Instance().RedisSentinelAddr,
	})
	if err := pingRedisClient(redisClient); err != nil {
		panic(err)
	}
	return redisClient
}

//func (m *Model) autoConnectRedisClient(tryTimes ,maxRetryTimes int) *redis.Client{
//
//}

func pingRedisClient(client *redis.Client) error {
	if pong, err := client.Ping().Result(); pong != "PONG" || err != nil {
		logger.Error("fail to ping redisClient", err)
		return err
	}
	return nil
}

func (m *Model) InitRedisCluster() *redis.ClusterClient {
	// clusterSlots 主要用于手动创建分布式集群
	//clusterSlots := func() ([]redis.ClusterSlot, error) {
	//	slots := []redis.ClusterSlot{
	//		// First node with 1 master and 1 slave.
	//		{
	//			Start: 0,
	//			End:   8191,
	//			Nodes: []redis.ClusterNode{{
	//				Addr: ":7000", // master
	//			}, {
	//				Addr: ":8000", // 1st slave
	//			}},
	//		},
	//		// Second node with 1 master and 1 slave.
	//		{
	//			Start: 8192,
	//			End:   16383,
	//			Nodes: []redis.ClusterNode{{
	//				Addr: ":7001", // master
	//			}, {
	//				Addr: ":8001", // 1st slave
	//			}},
	//		},
	//	}
	//	return slots, nil
	//}
	//redisClusterClient := redis.NewClusterClient(&redis.ClusterOptions{
	//	ClusterSlots:  clusterSlots,
	//	RouteRandomly: true, // 随机从主从节点路由
	//	RouteByLatency: false, // 从主从节点中延迟低的那个路由
	//})

	redisClusterClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          config.Instance().RedisClusterAddr,
		RouteByLatency: true,
	})
	if pong, err := redisClusterClient.Ping().Result(); pong != "PONG" || err != nil {
		logger.Error("fail to ping redisClusterClient", err)
	}
	err := redisClusterClient.ReloadState()
	if err != nil {
		panic(err)
	}
	return redisClusterClient
}

func (m *Model) GetStringWithDefaultValue(key string, defaultValue string) string {
	if err := pingRedisClient(m.RedisClient); err != nil {
		panic(err)
	}
	val := m.RedisClient.Get(key)
	if val.Err() == nil && val.Val() != "" {
		return val.Val()
	}
	return defaultValue
}

func (m *Model) SetStringValueWithoutExpireTime(key string, value interface{}) error {
	if err := pingRedisClient(m.RedisClient); err != nil {
		panic(err)
	}
	err := m.RedisClient.Set(key, value, 0).Err()
	return err
}
