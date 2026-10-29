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

The test server can be used manually as well:

```shell
go run ./cmd/wst-testsrv -c config.json
```

where the configuration is a JSON file with the keys documented in `cmd/wst-testsrv/main.go`. Special endpoints are
available under the `/_/` prefix (`ping`, `echo`, `status/<code>`, `delay`, `stream`, `stats`, `exit`); all other
paths are served from the configured script directory.
