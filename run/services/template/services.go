package template

import (
	"errors"
	"github.com/bukka/wst/run/services/template/service"
)

type Services map[string]service.TemplateService

func (svcs Services) Find(name string) (service.TemplateService, error) {
	svc, ok := svcs[name]
	if !ok {
		return nil, errors.New("service not found")
	}
	return svc, nil
}
