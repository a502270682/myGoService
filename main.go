package main

import (
	"context"
	logger "github.com/sirupsen/logrus"
	"myGoService/config"
	"myGoService/model"
	"myGoService/model/rabbitmq"
	"myGoService/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config.Init("./config/local_project.conf")
	s := service.NewService()
	rabbitmq.RabbitMessageQueueInit()
	s.Wf.Router.Run(config.Instance().LocalProjectPort)
	srv := &http.Server{
		Addr: config.Instance().LocalProjectPort,
		Handler: s.Wf.Router,
	}
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Kill, os.Interrupt, syscall.SIGBUS, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM)
	<-signals
	logger.Info("service start to shut down")

	ctx,cancel := context.WithTimeout(context.Background(),5*time.Second)
	defer cancel()
	defer model.CloseModelServer(s.Wf.M)
	if err := srv.Shutdown(ctx);err != nil {
		logger.Fatal("server shutdown failed:",err)
	}
	logger.Info("server shutDown success")
}
