package spec

import (
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"github.com/bukka/wst/run/instances"
	"github.com/bukka/wst/run/servers"
	"strings"
)

type Spec interface {
	ExecuteInstances(filteredInstances []string, dryRun bool) error
}

type Maker struct {
	env           app.Env
	instanceMaker *instances.InstanceMaker
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env:           env,
		instanceMaker: instances.CreateInstanceMaker(env),
	}
}

func (m *Maker) Make(config *types.Spec, servers servers.Servers) (Spec, error) {
	var instances []instances.Instance
	for _, instance := range config.Instances {
		inst, err := m.instanceMaker.Make(instance, servers)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}

	return &nativeSpec{
		workspace: config.Workspace,
		instances: instances,
	}, nil
}

type nativeSpec struct {
	workspace string
	instances []instances.Instance
}

func (n nativeSpec) ExecuteInstances(filteredInstances []string, dryRun bool) error {
	// Loop through the instances.
	for _, instance := range n.instances {
		// Determine the instance identifier or name.
		instanceName := instance.GetName()

		// Execute if filteredInstances is empty or nil, meaning execute all instances.
		if len(filteredInstances) == 0 {
			if err := instance.ExecuteActions(dryRun); err != nil {
				return err // Return immediately if any execution fails.
			}
		} else {
			// If filteredInstances is not empty, check if the instance name starts with any of the filteredInstances strings.
			for _, filter := range filteredInstances {
				if strings.HasPrefix(instanceName, filter) {
					// Execute the instance if it matches the filter.
					if err := instance.ExecuteActions(dryRun); err != nil {
						return err // Return immediately if any execution fails.
					}
					break // Move to the next instance after successful execution.
				}
			}
		}
	}

	return nil // Return nil if all selected instances were executed successfully.
}
