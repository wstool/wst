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

//go:build integration

// Package integration runs WST end to end against the wst-testsrv test server.
//
// Run with:
//
//	go test -tags integration ./integration/...
//
// Set WST_INTEGRATION_DEBUG=1 to get the WST debug log output.
package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/run"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	projectConfig string
	testsrvBinary string
)

func TestMain(m *testing.M) {
	code, err := setupAndRun(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "integration setup failed: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func setupAndRun(m *testing.M) (int, error) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		return 1, fmt.Errorf("cannot determine test file location")
	}
	integrationDir := filepath.Dir(thisFile)
	repoRoot := filepath.Dir(integrationDir)
	projectConfig = filepath.Join(integrationDir, "project", "wst.yaml")

	binDir, err := os.MkdirTemp("", "wst-integration-")
	if err != nil {
		return 1, err
	}
	defer os.RemoveAll(binDir)

	testsrvBinary = filepath.Join(binDir, "wst-testsrv")
	build := exec.Command("go", "build", "-o", testsrvBinary, "./cmd/wst-testsrv")
	build.Dir = repoRoot
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return 1, fmt.Errorf("building wst-testsrv: %w", err)
	}

	return m.Run(), nil
}

// runInstance executes a single instance of the integration project in-process, the same way the CLI does.
func runInstance(t *testing.T, name string) error {
	t.Helper()

	var logger *zap.Logger
	var err error
	if os.Getenv("WST_INTEGRATION_DEBUG") != "" {
		logger, err = zap.NewDevelopment()
	} else {
		cfg := zap.NewDevelopmentConfig()
		cfg.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
		logger, err = cfg.Build()
	}
	require.NoError(t, err)
	defer logger.Sync()

	fnd := app.NewFoundation(logger.Sugar(), false)
	options := &run.Options{
		ConfigPaths: []string{projectConfig},
		Overwrites: map[string]string{
			"spec.defaults.parameters.testsrv_binary": testsrvBinary,
		},
		NoEnvs:    true,
		Instances: []string{name},
	}
	return run.CreateRunner(fnd).Execute(options)
}

func TestLocalInstances(t *testing.T) {
	tests := []struct {
		name        string
		expectError bool
	}{
		{name: "basic"},
		{name: "request-body-slow"},
		{name: "request-body-incomplete"},
		{name: "fail-response-mismatch", expectError: true},
		{name: "fail-start", expectError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runInstance(t, tt.name)
			if tt.expectError {
				assert.Error(t, err, "instance %s is expected to fail", tt.name)
			} else {
				assert.NoError(t, err, "instance %s is expected to pass", tt.name)
			}
		})
	}
}
