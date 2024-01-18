package app

import (
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"os"
)

type Env struct {
	Logger *zap.SugaredLogger
	Fs     afero.Fs
}

func (env *Env) GetUserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (env *Env) LookupEnvVar(key string) (string, bool) {
	return os.LookupEnv(key)
}
