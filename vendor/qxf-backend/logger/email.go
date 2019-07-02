package logger

import (
	"fmt"
	"os"
	"qxf-backend/config"
	"qxf-backend/email"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	EMAIL_ROLE_ADMIN                  = "email_role_admin"                  // 管理员
	EMAIL_ROLE_NORMAL                 = "email_role_normal"                 // 普通(预留)
	EMAIL_ROLE_FUND                   = "email_role_fund"                   // 基金相关
	EMAIL_ROLE_CROWD_FUND             = "email_role_crowd_fund"             // 众筹相关
	EMAIL_ROLE_ORDER                  = "email_role_order"                  // 支付或者与订单相关
	EMAIL_ROLE_ACTIVITY               = "email_role_activity"               // 运营活动
	EMAIL_ROLE_EXT_CALL               = "email_role_ext_call"               // 外部调用的关注方
	EMAIL_ROLE_FUND_PM                = "email_role_fund_pm"                // 基金相关产品经理，包括开发
	EMAIL_ROLE_PRODUCT                = "email_role_product"                // 产品配置相关的运营和后台
	EMAIL_ROLE_PUSH                   = "email_role_push"                   // PUSH相关人员
	EMAIL_ROLE_FE_DEV                 = "email_role_fe_dev"                 // H5开发人员
	EMAIL_FOR_ONE_ROLE                = "email_role_for_one_%s"             // 当只给特定的一个用户发邮件时调用
	EMAIL_ROLE_PAYMENT_SERVICE        = "email_role_payment_service"        // 支付通道
	EMAIL_ROLE_PRODUCT_CENTER_SERVICE = "email_role_product_center_service" // 产品中心
)

var EmailRoleList = []string{
	EMAIL_ROLE_ADMIN,
	EMAIL_ROLE_NORMAL,
	EMAIL_ROLE_FUND,
	EMAIL_ROLE_CROWD_FUND,
	EMAIL_ROLE_ORDER,
	EMAIL_ROLE_ACTIVITY,
	EMAIL_ROLE_EXT_CALL,
	EMAIL_ROLE_FUND_PM,
	EMAIL_ROLE_PRODUCT,
	EMAIL_ROLE_FE_DEV,
	EMAIL_ROLE_PAYMENT_SERVICE,
	EMAIL_ROLE_PRODUCT_CENTER_SERVICE,
}

var EmailRolesMapLock = &sync.RWMutex{}
var EmailRolesMap = make(map[string]string)

type EmailBatchTool struct {
	sendChan         chan *subjectEmailSend
	emailContentChan chan *emailContent
	subjects         map[string]*subjectEmailSend
	health           bool
	delayTime        time.Duration
	channelSize      int
	maxSubJectNum    int
}

type subjectEmailSend struct {
	start   time.Time
	subject string
	count   int
	info    string
	isSend  bool
}

type emailContent struct {
	subject string
	info    string
}

func NewEmailBatchTool() *EmailBatchTool {
	//设置参数
	channelSize := 10000

	ret := &EmailBatchTool{
		sendChan:         make(chan *subjectEmailSend, channelSize),
		emailContentChan: make(chan *emailContent, channelSize),
		subjects:         map[string]*subjectEmailSend{},
		health:           true,
		delayTime:        time.Second * 60,
		channelSize:      channelSize,
		maxSubJectNum:    200,
	}
	go ret.sendBuf()
	go ret.mergeSubject()
	return ret
}

func (e *EmailBatchTool) FlushEmail() {
	e.health = false
}

func (e *EmailBatchTool) InsertEmail(subject, info string) {
	e.emailContentChan <- &emailContent{subject: subject, info: info}
}

//对email进行合并
func (e *EmailBatchTool) mergeSubject() {
	for {
		if !e.health {
			return
		}
		//获取需要合并的邮件
		tmp := <-e.emailContentChan
		subject := tmp.subject
		info := tmp.info

		//如果该subject不存在，则新建，并插入至sendChan中，在1分钟后发送
		if _, ok := e.subjects[subject]; !ok {
			e.subjects[subject] = &subjectEmailSend{subject: subject, count: 0, info: "", start: time.Now(), isSend: false}
			e.sendChan <- e.subjects[subject]
		}

		//如果该subject于1分钟前建立，该subject由sendBuf发送出去，且无需向其中写数据。新建并插入至sendChan中，在1分钟后发送
		if time.Now().After(e.subjects[subject].start.Add(e.delayTime)) {
			e.subjects[subject] = &subjectEmailSend{subject: subject, count: 0, info: "", start: time.Now(), isSend: false}
			e.sendChan <- e.subjects[subject]
		}
		ses := e.subjects[subject]

		ses.count++
		ses.info += "<p>" + info + "</p>"
		ses.info += "<p> 第" + strconv.Itoa(ses.count) + "个</p>"
		ses.info += "<p> -------------------------------------------------------------- </p>"

		//若subject总数到上限，马上发送【注：因为时间未到1分钟，可以保证sendBuf尚未发送】
		if ses.count >= e.maxSubJectNum {
			//标记已发送
			ses.isSend = true
			e.startSend(ses)
			delete(e.subjects, subject)
		}
	}
}

func (e *EmailBatchTool) startSend(ses *subjectEmailSend) {
	role := ses.subject[:strings.Index(ses.subject, "@")]
	subject := ses.subject[strings.Index(ses.subject, "@")+1:]
	go emailSend(subject+": "+strconv.Itoa(ses.count)+"条", ses.info, role)
}

//用于发送所有超过1分钟的邮件，或者把当前邮件全部发出
func (e *EmailBatchTool) sendBuf() {
	for {
		//获取最早的一封邮件
		ses := <-e.sendChan
		for {
			if !e.health {
				//全部发出
				for {
					if ses.isSend {
						continue
					}

					e.startSend(ses)
					ses = <-e.sendChan
				}
			}
			//在这里采用1分1秒，即可保证该subject不可再被修改
			if time.Now().After(ses.start.Add(e.delayTime+time.Second)) || ses.isSend {
				break
			}
			time.Sleep(time.Second)
		}
		if ses.isSend {
			continue
		}
		ses.isSend = true

		e.startSend(ses)
	}
}

func FlushEmail() {
	eb.FlushEmail()
}

func emailBatch(subject, info string, roles ...string) {
	if sendMailCount > MAXMAILCOUNT {
		return
	}
	//将需要发送的数据放入chan中，进行单线程处理
	if strings.Contains(subject, "LB Error: http") {
		subject = strings.Split(subject, "?")[0]
	}

	if len(roles) == 0 { //默认群组
		roles = []string{EMAIL_ROLE_ADMIN}
	}

	if config.Instance().Env == "staging" {
		Info("sending email to %s with subject: %s, info: %s", strings.Join(roles, ", "), subject, info)
	}

	for _, role := range roles {
		if _, ok := EmailRolesMap[role]; !ok {
			EmailRolesMapLock.Lock()
			key := fmt.Sprintf(EMAIL_FOR_ONE_ROLE, role)
			EmailRolesMap[key] = role + "@creditease.cn"
			EmailRolesMapLock.Unlock()
			eb.InsertEmail(key+"@"+subject, info)
		} else {
			eb.InsertEmail(role+"@"+subject, info)
		}
	}
}

func emailSend(subject, info string, role string) {
	if sendMailCount > MAXMAILCOUNT {
		return
	}

	roleReceivers := config.Instance().EmailReceiver
	if len(EmailRolesMap) > 0 {
		if roleList, ok := EmailRolesMap[role]; ok {
			roleReceivers = roleList
		}
	}

	go func() {
		m := email.Mail{
			From:    "noreply@bdp.yixin.com",
			To:      roleReceivers,
			Subject: subject,
		}
		m.HTML = "<p>" + strings.Replace(info, "\n", "<br/>", -1) + "</p>"
		m.HTML += "<p>" + config.Instance().IPAddr() + "</p>"
		m.HTML += "<p>" + time.Now().Format("2006-01-02 15:04:05") + "</p>"
		err := m.Send(config.Instance().SMTPHost, 25, config.Instance().SMTPUser, config.Instance().SMTPPassword)
		if err != nil {
			Warn("send email failed: %v", err)
		}
	}()
	sendMailCount++
	if sendMailCount > MAXMAILCOUNT {
		go func() {
			m := email.Mail{
				From:    "noreply@bdp.yixin.com",
				To:      roleReceivers,
				Subject: "今日发送报警邮件已到上限，将不再发送",
			}
			m.HTML = "<p>" + "今日发送报警邮件已到上限，将不再发送" + "</p>"
			m.HTML += "<p>" + config.Instance().IPAddr() + "</p>"
			m.HTML += "<p>" + time.Now().Format("2006-01-02 15:04:05") + "</p>"
			err := m.Send(config.Instance().SMTPHost, 25, config.Instance().SMTPUser, config.Instance().SMTPPassword)
			if err != nil {
				Warn("send email failed: %v", err)
			}
		}()
	}
}

// 获得在当前的应用名称
func getLainAppName() string {
	app, ok := os.LookupEnv("LAIN_APPNAME")
	if ok {
		return app
	}
	return ""
}

func getEmailBatchLainAppNameSuffix() string {
	appNameSuffix := getLainAppName()
	if appNameSuffix != "" {
		appNameSuffix = fmt.Sprintf(" from %s", appNameSuffix)
	}
	return appNameSuffix
}

// 以下Email方法如不传入roles，默认发给Admin
// 所有roles均包含了Admin成员，即如只传Activity会发给Admin和Activity的所有成员
//低频、尽快人工介入
func EmailBatchUrgent(subject, info string, roles ...string) {
	emailBatch("【紧急】"+subject+getEmailBatchLainAppNameSuffix(), info, roles...)
}

//需要人工介入，当天处理
func EmailBatchImportant(subject, info string, roles ...string) {
	emailBatch("【重要】"+subject+getEmailBatchLainAppNameSuffix(), info, roles...)
}

//不需要人工介入，只需要观察，如周末报警推迟到下一工作日处理
func EmailBatchWarn(subject, info string, roles ...string) {
	emailBatch("【警告】"+subject+getEmailBatchLainAppNameSuffix(), info, roles...)
}

//普通
func EmailBatchNormal(subject, info string, roles ...string) {
	emailBatch("【普通】"+subject+getEmailBatchLainAppNameSuffix(), info, roles...)
}
