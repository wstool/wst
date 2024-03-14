# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### App
- possibly some other helpers

### Run
- configs saving to local service on start
- loading server templates (custom template function)
- docker environment integration
- kubernetes environment integration
- dry run mode - integrate to action, services and sandboxes
- load testing action (wrk runner)
- tests

### Config
- local env port should be start / end instead of from / to maybe
- add checks for casting to lower integers (e.g. int16)
- factories
- merging - generic merging rules taken from params

## Plan

- [x] March 1: Template rendering
- [ ] March 2: Docker and Kubernetes envs
- [ ] March 3: Dry run and finalization of all envs
- [ ] March 4: Benchmarking actions (wrk)
- [ ] April 1: Config factories, clean up and extra checks and parser tests
- [ ] April 2: Config merging and finalize tests
- [ ] April 3: Actions tests
- [ ] April 4: Environments tests
- [ ] May 1: Instance, Spec and Parameters tests
- [ ] May 2: Resources and sandboxes tests and clean up (move merging logic)
- [ ] May 3: Servers tests
- [ ] May 4: Services tests
- [ ] June 1: Run tests
- [ ] June 2: Full tests