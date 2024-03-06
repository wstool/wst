package template

type Config struct {
	User  string
	Group string
	Dirs  struct {
		Run    string
		Script string
	}
}

type Service struct {
	Address string
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
