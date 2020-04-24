package rabbitmq

import (
	"context"
	"fmt"
	logger "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"time"
)

const CreatedAt = "created_at"

func (mq *MessageQueueConnect) ReceiveFromRabbitMq(ctx context.Context, queueName string,
	callback func(ctx context.Context, msg amqp.Delivery) error) error {
	//var (
	//	msg amqp.Delivery
	//	ok bool
	//	mqch *amqp.Channel
	//)

	mqch, err := mq.Conn.Channel()
	if err != nil {
		time.Sleep(mq.SleepTime)
		return err
	}

	del := make([]amqp.Delivery, 0)
	for i := 0; i < 30; i++ {
		msg, ok, err := mqch.Get(queueName, false)
		if err != nil {
			break
		}
		if ok {
			if tmp, ok := msg.Headers[CreatedAt]; !ok || tmp == nil {
				msg.Headers = amqp.Table{CreatedAt: time.Now().Unix()}
			}
			del = append(del, msg)
		} else {
			break
		}
	}
	for _, msg := range del {
		mq.handleMsg(ctx, queueName, msg, callback)
	}
	mqch.Close()
	return nil
}

func (mq *MessageQueueConnect) write(exchange, routingKey, exchangeType string, body []byte, headers amqp.Table) error {
	mqch, err := mq.Conn.Channel()
	defer mqch.Close()
	/*
		带绑定版本
		// 验证目标队列是否存在
		_, err = mqch.QueueDeclarePassive("queueName", true, false, false, true, nil)
		if err != nil {
			return err
			// 队列不存在,声明队列
			// name:队列名称;durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;noWait:是否非阻塞,
			// true为是,不等待RMQ返回信息;args:参数,传nil即可;exclusive:是否设置排他
			//_, err := mqch.QueueDeclare("queueName", true, false, false, true, nil)
			//if err != nil {
			//	logger.Error("fail to register queue, err", err)
			//	return err
			//}
		}
		// 队列绑定
		err = mqch.QueueBind("queueName",routingKey,exchange,true,nil)
		if err != nil {
			logger.Error("fail to queueBind, err",err)
			return err
		}
		// 用于检查交换机是否存在,已经存在不需要重复声明
		err = mqch.ExchangeDeclarePassive(exchange,exchangeType,true, false, false, true, nil)
		if err != nil {
			logger.Error("fail to ExchangeDeclarePassive, err",err)
			return err
			// 注册交换机
			// name:交换机名称,kind:交换机类型,durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;
			// noWait:是否非阻塞, true为是,不等待RMQ返回信息;args:参数,传nil即可; internal:是否为内部
			//err = mqch.ExchangeDeclare(exchange,exchangeType,true,false,false,true,nil)
			//if err != nil {
			//	return err
			//}
		}
	*/

	err = mqch.Publish(
		exchange,
		routingKey,
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Headers:      headers,
			Body:         body,
		})
	return err
}

func (mq *MessageQueueConnect) WriteData(ctx context.Context, exchange, routingKey, exchangeType string, msg amqp.Delivery) error {
	if tmp, ok := msg.Headers[CreatedAt]; !ok || tmp == nil {
		// 将当前时间写入
		msg.Headers = amqp.Table{CreatedAt: time.Now().Unix()}
	}
	return mq.write(exchange, routingKey, exchangeType, msg.Body, msg.Headers)
}

func (mq *MessageQueueConnect) handleMsg(ctx context.Context, queueName string, msg amqp.Delivery,
	callBack func(ctx context.Context, msg amqp.Delivery) error) {
	// callBack 系业务处理代码
	err := callBack(ctx, msg)
	if err != nil {
		err = mq.WriteData(ctx, "", queueName, "", msg)
		if err != nil {
			logger.Error(fmt.Sprintf("fail to WriteData to mq , queueName :%s", queueName), err)
		}
		if err = confirmOne(ctx, msg); err != nil {
			return
		}
	} else {
		confirmOne(ctx, msg)
	}
}

func confirmOne(ctx context.Context, msg amqp.Delivery) error {
	err := msg.Ack(false)
	if err != nil {
		logger.Error("fail to Ack msg, err :%+v", err)
		return err
	}
	return nil
}
