package buslogic

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	logger "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"myGoService/code"
	"myGoService/config"
	"myGoService/model/rabbitmq"
	"time"
)

type LoginRequest struct {
	UserId   int64  `form:"user_id" binding:"required"`
	Password string `form:"password" binding:"required"`
	Name     string `form:"name" binding:"required"`
}

type LoginResponse struct {
}

func (wf *WorkFlow) GetLogin(ctx *gin.Context) {
	lr := LoginRequest{}
	if err := ctx.ShouldBindWith(&lr, binding.Query); err != nil {
		logger.Error("登陆请求参数缺失或错误", err)
		wf.ResponseCode2OutputMessage(ctx, code.RequestInputParamsMissingOrError)
		return
	}
	logger.Info(wf.M.GetStringWithDefaultValue("test", ""))
	err := wf.CheckUserLoginPassword(&lr)
	if err != nil {
		logger.Error("登陆验证密码失败", err)
		wf.ResponseCode2OutputMessage(ctx, code.UserPasswordWrongOrMissing)
	} else {
		logger.Info("登陆成功")
		wf.ResponseCode2OutputMessage(ctx, code.RequestOk)
	}
	// 将登陆信息写入mq
	mq := rabbitmq.GetMQC()
	mq.WriteData(ctx, config.Instance().RMQExchangeLogin.RMQExchangeLoginName, config.Instance().RMQExchangeLogin.RMQExchangeLoginRoutingKeyInfo, "", amqp.Delivery{
		Body: []byte(lr.Name),
		Headers: amqp.Table{
			rabbitmq.CreatedAt: time.Now()}})
}

func (wf *WorkFlow) GetLoginMessageQueue(ctx *gin.Context) {
	mq := rabbitmq.GetMQC()
	callBack := func(ctx context.Context, msg amqp.Delivery) error {
		logger.Info("the receive msg : %+v", string(msg.Body))
		return nil
	}
	err := mq.ReceiveFromRabbitMq(ctx, config.Instance().RMQLoginQueue, callBack)
	if err != nil {
		logger.Error("fail to ReceiveFromRabbitMq,err :%+v", err)
	}
}

func (wf *WorkFlow) CheckUserLoginPassword(lr *LoginRequest) error {
	user, err := wf.M.GetUserById(lr.UserId)
	if err != nil {
		return errors.New("获取用户信息失败")
	}
	if user.Password != lr.Password {
		return errors.New("用户验证密码失败")
	}
	return nil
}
