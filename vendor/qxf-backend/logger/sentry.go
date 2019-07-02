package logger

import (
	"github.com/kusora/raven-go"
	"qxf-backend/config"
)

func InitSentry(app string) {
	raven.SetDSN(config.Instance().DSN)
	raven.SetTagsContext(map[string]string{"ip_addr": config.Instance().IPAddr(), "app": app})
}
