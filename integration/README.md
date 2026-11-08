# Integration tests

The integration tests run WST end to end against `wst-testsrv`, a small purpose-built HTTP server located in
`cmd/wst-testsrv`. The server is deterministic and dependency free, so the same instances can run locally, in Docker
and in Kubernetes without installing third-party servers.

The WST project used by the tests lives in `project/`. Each file in `project/spec/instances/` is one instance and the
Go test in `integration_test.go` runs the selected instances in-process through the runner, the same way the CLI does.
Instances whose name starts with `fail-` are negative tests that are expected to fail.

Run the tests with:

```shell
go test -tags integration ./integration/...
```

Set `WST_INTEGRATION_DEBUG=1` to get the full WST debug log.

## Kubernetes

The same instances can be run with the `kubernetes` sandbox. The test setup builds a static `wst-testsrv` binary,
creates the `wst-testsrv:local` image from `cmd/wst-testsrv/Dockerfile` and optionally loads it to the cluster. The
Kubernetes tier is opt-in and configured with environment variables:

| Variable                       | Description                                                                                                                                                                |
|--------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `WST_INTEGRATION_KUBERNETES`   | Set to `1` to run the Kubernetes instances.                                                                                                                                |
| `WST_INTEGRATION_KUBECONFIG`   | Kubeconfig path. Defaults to `$KUBECONFIG` or `~/.kube/config`.                                                                                                            |
| `WST_INTEGRATION_NAMESPACE`    | Namespace for the created objects. Defaults to `default`.                                                                                                                  |
| `WST_INTEGRATION_IMAGE_LOADER` | How to make the image available in the cluster: `minikube` (`minikube image load`), `kind` (`kind load docker-image`) or empty when the cluster already sees local images. |

For example with minikube:

```shell
minikube start
WST_INTEGRATION_KUBERNETES=1 WST_INTEGRATION_IMAGE_LOADER=minikube go test -tags integration ./integration/...
```

and with kind:

```shell
kind create cluster
WST_INTEGRATION_KUBERNETES=1 WST_INTEGRATION_IMAGE_LOADER=kind go test -tags integration ./integration/...
```

Public services are reached through a port forward from the environment port on `127.0.0.1` to the pod, so no
load balancer or node port access is needed.

## Test server

The test server can be used manually as well:

```shell
go run ./cmd/wst-testsrv -c config.json
```

where the configuration is a JSON file with the keys documented in `cmd/wst-testsrv/main.go`. Special endpoints are
available under the `/_/` prefix (`ping`, `echo`, `status/<code>`, `delay`, `stream`, `upload`, `stats`, `exit`);
all other paths are served from the configured script directory. The `echo` endpoint reports how the request body
arrived (`body_chunks`, `body_read_ms`) and the `upload` endpoint behaves like a server with a request body timeout
(`timeout_ms`), responding with 408 when the body stalls and 400 when the client disconnects.
