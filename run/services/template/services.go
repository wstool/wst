package template

import "github.com/bukka/wst/run/services"

type Service struct {
	service services.Service
}

func (s *Service) Address() (string, error) {
	return s.service.BaseUrl()
}

func (s *Service) Dirs() map[string]string {
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
