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
// The local instances always run. The container tiers are opt-in through environment variables:
//
//	WST_INTEGRATION_KUBERNETES=1      run the instances in Kubernetes as well
//	WST_INTEGRATION_KUBECONFIG=<path> kubeconfig to use (defaults to $KUBECONFIG or ~/.kube/config)
//	WST_INTEGRATION_NAMESPACE=<ns>    Kubernetes namespace to use (defaults to "default")
//	WST_INTEGRATION_IMAGE_LOADER=<l>  how to make the wst-testsrv:local image available to the cluster:
//	                                  "minikube" (minikube image load), "kind" (kind load docker-image) or
//	                                  empty when the cluster can already see the local Docker images
//	WST_INTEGRATION_DEBUG=1           get the WST debug log output
package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/wstool/wst/app"
	"github.com/wstool/wst/run"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// testsrvImage is the image name used by the container sandboxes in the integration project.
	testsrvImage = "wst-testsrv:local"
	// testsrvContainerBinary is the path of the server binary inside the image (see cmd/wst-testsrv/Dockerfile).
	testsrvContainerBinary = "/usr/local/bin/wst-testsrv"
)

var (
	repoRoot      string
	projectConfig string
	testsrvBinary string
	// instances lists all instances of the integration project that are run in every tier.
	instances = []struct {
		name        string
		expectError bool
	}{
		{name: "basic"},
		{name: "request-body-slow"},
		{name: "request-body-incomplete"},
		{name: "fail-response-mismatch", expectError: true},
		{name: "fail-start", expectError: true},
	}
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
	repoRoot = filepath.Dir(integrationDir)
	projectConfig = filepath.Join(integrationDir, "project", "wst.yaml")

	binDir, err := os.MkdirTemp("", "wst-integration-")
	if err != nil {
		return 1, err
	}
	defer os.RemoveAll(binDir)

	testsrvBinary = filepath.Join(binDir, "wst-testsrv")
	// The binary is built statically so that the same file can be copied to the container image.
	if err := runCommand(repoRoot, []string{"CGO_ENABLED=0"},
		"go", "build", "-o", testsrvBinary, "./cmd/wst-testsrv"); err != nil {
		return 1, fmt.Errorf("building wst-testsrv: %w", err)
	}

	if kubernetesEnabled() {
		if err := buildContainerImage(binDir); err != nil {
			return 1, err
		}
		if err := loadContainerImage(); err != nil {
			return 1, err
		}
	}

	return m.Run(), nil
}

func runCommand(dir string, env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	fmt.Fprintf(os.Stderr, "integration setup: running %s %s\n", name, strings.Join(args, " "))
	return cmd.Run()
}

// buildContainerImage builds the wst-testsrv:local image from the freshly built static binary.
func buildContainerImage(binDir string) error {
	dockerfile := filepath.Join(repoRoot, "cmd", "wst-testsrv", "Dockerfile")
	if err := runCommand(binDir, nil, "docker", "build", "-t", testsrvImage, "-f", dockerfile, binDir); err != nil {
		return fmt.Errorf("building %s image: %w", testsrvImage, err)
	}
	return nil
}

// loadContainerImage makes the image available inside the Kubernetes cluster.
func loadContainerImage() error {
	loader := os.Getenv("WST_INTEGRATION_IMAGE_LOADER")
	var err error
	switch loader {
	case "":
		return nil
	case "minikube":
		err = runCommand(repoRoot, nil, "minikube", "image", "load", "--overwrite=true", testsrvImage)
	case "kind":
		err = runCommand(repoRoot, nil, "kind", "load", "docker-image", testsrvImage)
	default:
		return fmt.Errorf("unknown WST_INTEGRATION_IMAGE_LOADER value %q", loader)
	}
	if err != nil {
		return fmt.Errorf("loading %s image with %s: %w", testsrvImage, loader, err)
	}
	return nil
}

func kubernetesEnabled() bool {
	return os.Getenv("WST_INTEGRATION_KUBERNETES") != ""
}

func kubeconfigPath(t *testing.T) string {
	t.Helper()
	if path := os.Getenv("WST_INTEGRATION_KUBECONFIG"); path != "" {
		return path
	}
	if path := os.Getenv("KUBECONFIG"); path != "" {
		return path
	}
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	return filepath.Join(home, ".kube", "config")
}

func kubernetesNamespace() string {
	if ns := os.Getenv("WST_INTEGRATION_NAMESPACE"); ns != "" {
		return ns
	}
	return "default"
}

// runInstance executes a single instance of the integration project in-process, the same way the CLI does.
func runInstance(t *testing.T, name string, overwrites map[string]string) error {
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
		Overwrites:  overwrites,
		NoEnvs:      true,
		Instances:   []string{name},
	}
	return run.CreateRunner(fnd).Execute(options)
}

func runInstances(t *testing.T, overwrites map[string]string) {
	t.Helper()
	for _, tt := range instances {
		t.Run(tt.name, func(t *testing.T) {
			err := runInstance(t, tt.name, overwrites)
			if tt.expectError {
				assert.Error(t, err, "instance %s is expected to fail", tt.name)
			} else {
				assert.NoError(t, err, "instance %s is expected to pass", tt.name)
			}
		})
	}
}

func TestLocalInstances(t *testing.T) {
	runInstances(t, map[string]string{
		"spec.defaults.parameters.testsrv_binary": testsrvBinary,
	})
}

func TestKubernetesInstances(t *testing.T) {
	if !kubernetesEnabled() {
		t.Skip("set WST_INTEGRATION_KUBERNETES=1 to run the Kubernetes integration tests")
	}
	runInstances(t, map[string]string{
		"spec.defaults.parameters.testsrv_binary": testsrvContainerBinary,
		"spec.defaults.service.sandbox":           "kubernetes",
		"spec.environments.kubernetes.kubeconfig": kubeconfigPath(t),
		"spec.environments.kubernetes.namespace":  kubernetesNamespace(),
	})
}
