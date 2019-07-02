package model

import (
	"github.com/go-xorm/xorm"
	"goServices/config"
	_ "github.com/go-sql-driver/mysql"
	"qxf-backend/logger"
	"time"
)


func (m *Model) InitMysql()  {
	if m.MysqlClient == nil {
		m.MysqlClient = connectMysql()
	}
	if m.MysqlClient.Ping() != nil {
		m.autoConnectMysql(0, 5)
	}
}

func connectMysql() *xorm.Engine {
	mysqlClient, err := xorm.NewEngine("mysql", config.Instance().MysqlConnDsn)
	if err != nil {
		logger.Error("fail to connectMysql", err)
		return nil
	}
	mysqlClient.SetMaxIdleConns(3)
	mysqlClient.SetMaxOpenConns(20)
	mysqlClient.SetConnMaxLifetime(0)
	return mysqlClient
}

func (m *Model) autoConnectMysql(tryTimes, maxRetryTimes int) {
	for tryTimes < maxRetryTimes {
		if err := m.MysqlClient.Ping(); err != nil {
			logger.Error("fail to Ping mysqlClient", err)
			time.Sleep(5 * time.Second)
			tryTimes++
			m.autoConnectMysql(tryTimes, maxRetryTimes)
		}
	}
}
