package model

import (
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	logger "github.com/sirupsen/logrus"
	"myGoService/config"
)

type Model struct {
	MysqlClient        *xorm.Engine
	RedisClient        *redis.Client
	RedisClusterClient *redis.ClusterClient
}

func NewModel() *Model {
	ret := new(Model)
	logger.Info("start to init model server!")
	ret.InitMysql()
	ret.InitRedisClient(config.Instance().RedisDbDefault)
	logger.Info("finish init model server!")
	return ret
}

func CloseModelServer(model *Model) {
	err := model.MysqlClient.Close()
	if err != nil {
		logger.Error("fail to close mysql, err:", err)
	}
	err = model.RedisClient.Close()
	if err != nil {
		logger.Error("fail to close redis, err:", err)
	}
	err = model.RedisClusterClient.Close()
	if err != nil {
		logger.Error("fail to close redis cluster, err:", err)
	}
}
