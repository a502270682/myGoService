package buslogic

import (
	"github.com/gin-gonic/gin"
	"myGoService/model"
	"myGoService/code"
)

type WorkFlow struct {
	M      *model.Model
	Router *gin.Engine
}

func NewWorkFlow() *WorkFlow {
	wf := new(WorkFlow)
	router := gin.Default()
	//router.Use(gin.Recovery())
	wf.Router = router
	wf.RegisterHandler()
	wf.M = model.NewModel()
	return wf
}

func (wf *WorkFlow) RegisterHandler() {
	// 分组1
	v1 := wf.Router.Group("/v1")
	v1.GET("healthCheck", func(context *gin.Context) {
		context.JSON(code.RequestOk,gin.H{
			"return_code": 0,
			"return_msg": "alive",
		})
	})
	v1.POST("login",wf.GetLogin)
	v1.POST("login/message/queue",wf.GetLoginMessageQueue)
}

func (wf *WorkFlow) ResponseCode2OutputMessage(context *gin.Context, responseCode int) {
	var responseMsg string
	switch responseCode {
	case code.RequestOk:
		responseMsg = "请求成功"
	case code.RequestInputParamsMissingOrError:
		responseMsg = "请求参数缺失或错误"
	case code.UserPasswordWrongOrMissing:
		responseMsg = "用户验证码为空或错误"
	case code.ErrorUnKnown:
		responseMsg = "未知错误"
	}
	header := gin.H{
		"message":       responseMsg,
		"response_code": responseCode,
	}
	context.JSON(200, header)
	return
}