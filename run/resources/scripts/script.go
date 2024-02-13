package scripts

import (
	"fmt"
	"github.com/bukka/wst/app"
	"github.com/bukka/wst/conf/types"
	"os"
	"strconv"
)

type Script interface {
}

type Scripts map[string]Script

type Maker struct {
	env app.Env
}

func CreateMaker(env app.Env) *Maker {
	return &Maker{
		env: env,
	}
}

func (m *Maker) Make(config map[string]types.Script) (Scripts, error) {
	scripts := make(Scripts)
	for scriptName, scriptConfig := range config {
		mode, err := strconv.ParseUint(scriptConfig.Mode, 8, 32)
		if err != nil {
			return nil, fmt.Errorf("error parsing file mode for script %s: %v", scriptName, err)
		}
		script := &nativeScript{
			content: scriptConfig.Content,
			path:    scriptConfig.Path,
			mode:    os.FileMode(mode),
		}
		scripts[scriptName] = script
	}
	return scripts, nil
}

type nativeScript struct {
	content string
	path    string
	mode    os.FileMode
}
