package main

import (
	"os"
	"os/signal"
	"syscall"
	"qxf-backend/logger"
	"goServices/service"
	"context"
	"time"
	"net/http"
	"goServices/model"
	"goServices/config"
	"goServices/model/rabbitmq"
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
	signal.Notify(signals, os.Kill, os.Interrupt, syscall.SIGBUS, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGTERM)
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
