// Copyright 2024 Jakub Zelenka and The WST Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package external

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type MockLogger struct {
	SugaredLogger *zap.SugaredLogger
	ObservedLogs  *observer.ObservedLogs
}

func NewMockLogger() *MockLogger {
	core, observedLogs := observer.New(zap.DebugLevel)
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
