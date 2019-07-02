package rabbitmq

import (
	"github.com/streadway/amqp"
	"myGoService/config"
	"qxf-backend/logger"
	"sync"
	"time"
)

const (
	// 交换器
	MQExchangeLoginQueue = "lp_login_queue"
	LoginQueueRountingKeyInfo = "info"
	LoginQueueName = "lp_login_test"
)

var (
	maxMqLength int
	count       int
	mux         = sync.RWMutex{}
	mqc         []*MessageQueueConnect
)

type MessageQueueConnect struct {
	Addr      string
	Conn      *amqp.Connection
	Lock      *sync.RWMutex
	SleepTime time.Duration
}

//初始化连接池
func RabbitMessageQueueInit() {
	setAddr()
	maxMqLength = len(mqc)
	for i := 0; i < maxMqLength; i++ {
		_, err := mqc[i].getConnectMq()
		if err != nil {
			panic(err)
		}
	}
}

//设置mqc连接池
func setAddr() {
	maxMqLength = len(config.Instance().RabbitMqAddrs)
	mux.Lock()
	defer mux.Unlock()
	tmp := make([]*MessageQueueConnect, 0)
	for i := 0; i < maxMqLength; i++ {
		tmp = append(tmp, &MessageQueueConnect{
			Addr: config.Instance().RabbitMqAddrs[i],
		})
	}
	mqc = tmp
}

func GetMQC() *MessageQueueConnect{
	mux.Lock()
	defer mux.Unlock()
	count ++
	if count > 9999 {
		count = 0
	}
	return mqc[count%maxMqLength]
}

// 单个mq连接
func (mq *MessageQueueConnect) getConnectMq() (*MessageQueueConnect, error) {
	var err error
	mq.Conn, err = amqp.Dial(mq.Addr)
	if err != nil {
		logger.Info("MessageQueue", "重新连接......")
		time.Sleep(time.Second * 60)
		setAddr()
		mq.getConnectMq()
		logger.Error("MessageQueue--->", "连接失败")
	} else {
		logger.Info("MessageQueue", "连接成功......")
		// 监听连接关闭
		go mq.confirmClose(mq.Conn.NotifyClose(make(chan *amqp.Error, 1)))
	}
	//defer mq.Close()
	return mq, err
}

func (mq *MessageQueueConnect) confirmClose(closeNotify <-chan *amqp.Error) {
	select {
	case confirmed := <-closeNotify:
		message := "客户端连接关闭"
		if confirmed.Server {
			message = "服务器端连接关闭"
		}
		mq.getConnectMq()
		logger.Info("MessageQueue", message, "code", confirmed.Code, "reason", confirmed.Reason)
	}
}

func (mq *MessageQueueConnect) SetSleepTime(t time.Duration) {
	mq.SleepTime = t
}

