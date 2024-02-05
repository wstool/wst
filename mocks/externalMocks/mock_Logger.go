package externalMocks

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type MockLogger struct {
	SugaredLogger *zap.SugaredLogger
	ObservedLogs  *observer.ObservedLogs
}

func NewMockLogger() *MockLogger {
	core, observedLogs := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	return &MockLogger{
		SugaredLogger: logger.Sugar(),
		ObservedLogs:  observedLogs,
	}
}

func (l *MockLogger) Messages() []string {
	var messages []string
	for _, log := range l.ObservedLogs.All() {
		messages = append(messages, log.Message)
	}
	return messages
}
