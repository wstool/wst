# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### App
- possibly some other helpers

### Run

- actions tests
- metrics tests
- expectations tests
- environments tests
- instance tests
- spec tests
- parameters tests
- resources tests
- sandboxes tests
- servers tests
  - specifically check setting of user and group as well as inheriting of the port
- services tests
- run tests
- integration tests
- add detailed info and debug logging
- kubernetes environment improvements
  - add health probes setup
- docker environment improvements
  - health check - waiting for container to be able to serve the traffic
  - custom docker 
- support metrics server expectation

### Config

- merging - generic merging rules taken from params
- boolean conversion in the same way as string (maybe something more generic that can handle both)

## Plan

- [x] Template rendering
- [x] Docker and Kubernetes envs
- [x] Dry run and finalization of all envs
- [x] Benchmarking actions (wrk)
- [x] Config factories, clean up and extra checks and parser tests
- [x] Config merging and finalize tests
- [x] Config overwrites and resolving all build issues
- [ ] Actions, expectations and metrics tests
- [ ] Environments tests
- [ ] Instance, Spec and Parameters tests
- [ ] Resources and sandboxes tests and clean up (move merging logic)
- [ ] Servers tests
- [ ] Services tests
- [ ] Run tests
- [ ] WST integration tests