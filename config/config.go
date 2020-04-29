package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	/*-----------------MYSQL------------------*/
	MYSQL_DSN = "root:@tcp(127.0.0.1:3306)/localproject?charset=utf8"

	/*-----------------REDIS------------------*/
	REDIS_ADDR       = "127.0.0.1:6379"
	DEFAULT_REDIS_DB = 0
)

var conf *Config

type Config struct {
	Port string `json:"port"`
	Env  string `json:"env"`
	// mysql
	MysqlConnDsn string `json:"mysql_conn_dsn"`
	// redis
	RedisAddr         string   `json:"redis_addr"`
	RedisClusterAddr  []string `json:"redis_cluster_addr"`
	RedisDbDefault    int      `json:"redis_db_default"`
	RedisMasterName   string   `json:"redis_master_name"`
	RedisSentinelAddr []string `json:"redis_sentinel_addr"`
	// mq
	RabbitMqAddrs    []string `json:"rabbit_mq_addrs"`
	RabbitMqPoolSize int      `json:"rabbit_mq_pool_size"`
	RMQExchangeLogin struct { // 登陆交换器
		RMQExchangeLoginName           string `json:"rmq_exchange_login_name"`
		RMQExchangeLoginRoutingKeyInfo string `json:"rmq_exchange_login_routing_key_info"`
	} `json:"rmq_exchange_login"`
	RMQLoginQueue string `json:"rmq_login_queue"`
}

func Init(path string) *Config {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalln(fmt.Sprintf("load conf %+v failed:", path), err)
	}
	err = json.Unmarshal(buf, &conf)
	if err != nil {
		log.Fatalln("decode conf file failed:", string(buf), err)
	}
	return conf
}

func Instance() *Config {
	if conf == nil {
		Init("./config/local_project.conf")
	}
	return conf
}
