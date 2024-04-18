package template

import (
	"errors"
	"github.com/bukka/wst/run/sandboxes/sandbox"
)

type Service interface {
	PrivateUrl() (string, error)
	Pid() (int, error)
	Dirs() map[sandbox.DirType]string
	Group() string
	User() string
	EnvironmentConfigPaths() map[string]string
}

type Services map[string]Service

func (svcs Services) Find(name string) (Service, error) {
	svc, ok := svcs[name]
	if !ok {
		return nil, errors.New("service not found")
	}
	return svc, nil
}
