package template

import (
	"github.com/bukka/wst/run/sandboxes/sandbox"
	"github.com/bukka/wst/run/services"
)

type Service struct {
	service services.Service
}

func NewService(service services.Service) *Service {
	return &Service{service: service}
}

func (s *Service) Address() (string, error) {
	return s.service.PrivateUrl()
}

func (s *Service) Pid() (int, error) {
	return s.service.Pid()
}

func (s *Service) Dirs() map[sandbox.DirType]string {
	return s.service.Dirs()
}

func (s *Service) Group() string {
	return s.service.Group()
}

func (s *Service) User() string {
	return s.service.User()
}

type Services struct {
	services services.Services
}

func (s *Services) Find(name string) Service {
	return Service{service: s.services[name]}
}

func NewServices(services services.Services) *Services {
	return &Services{services: services}
}
