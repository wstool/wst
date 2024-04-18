# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### App
- possibly some other helpers

### Run

- actions tests
- environments integration tests
- instance tests
- spec tests
- parameters tests
- resources tests
- sandboxes tests
- servers tests
  - specifically check setting of user and group as well as inheriting of the port
- services tests
- run tests
- add detailed info and debug logging
- kubernetes environment improvements
  - add health probes setup
- docker environment improvements
  - health check - waiting for container to be able to serve the traffic

### Config

- merging - generic merging rules taken from params
- boolean conversion in the same way as string (maybe something more generic that can handle both)

## Plan

- [x] March 1: Template rendering
- [x] March 2: Docker and Kubernetes envs
- [x] March 3: Dry run and finalization of all envs
- [x] March 4: Benchmarking actions (wrk)
- [x] April 1: Config factories, clean up and extra checks and parser tests
- [x] April 2: Config merging and finalize tests
- [x] April 3: Config overwrites and resolving all build issues
- [ ] April 4: Actions tests
- [ ] May 1: Environments tests
- [ ] May 2: Instance, Spec and Parameters tests
- [ ] May 3: Resources and sandboxes tests and clean up (move merging logic)
- [ ] May 4: Servers tests
- [ ] June 1: Services tests
- [ ] June 2: Full tests