package expect

import (
	"fmt"
	"github.com/bukka/wst/services"
)

type ExpectationAction struct {
	Service services.Service
}

func getService(svcs services.Services, name string) (services.Service, error) {
	svc, ok := svcs[name]
	if !ok {
		return svc, fmt.Errorf("service %s not found", name)
	}
	return svc, nil
}
