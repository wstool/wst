package template

import "github.com/bukka/wst/run/services"

type Service struct {
	service services.Service
}

func (s *Service) Address() (string, error) {
	return s.service.BaseUrl()
}

type Services struct {
	services map[string]Service
}

func (s *Services) Find(name string) Service {
	return s.services[name]
}

func NewServices(services map[string]Service) *Services {
	return &Services{services: services}
}
