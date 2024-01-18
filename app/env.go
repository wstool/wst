package app

import (
	"github.com/spf13/afero"
	"go.uber.org/zap"
	"os"
)

type Env interface {
	Logger() *zap.SugaredLogger
	Fs() afero.Fs
	GetUserHomeDir() (string, error)
	LookupEnvVar(key string) (string, bool)
}

type DefaultEnv struct {
	logger *zap.SugaredLogger
	fs     afero.Fs
}

func CreateEnv(logger *zap.SugaredLogger, fs afero.Fs) Env {
	return &DefaultEnv{
		logger: logger,
		fs:     fs,
	}
}

func (e *DefaultEnv) Logger() *zap.SugaredLogger {
	return e.logger
}

func (e *DefaultEnv) Fs() afero.Fs {
	return e.fs
}

func (e *DefaultEnv) GetUserHomeDir() (string, error) {
	return os.UserHomeDir()
}

func (e *DefaultEnv) LookupEnvVar(key string) (string, bool) {
	return os.LookupEnv(key)
}
