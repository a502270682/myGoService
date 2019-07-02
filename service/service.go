package service

import (
	"myGoService/buslogic"
)

type Service struct {
	Wf *buslogic.WorkFlow
}

func NewService() *Service {
	service := &Service{
		Wf: buslogic.NewWorkFlow(),
	}
	return service
}
