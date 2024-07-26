# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### App
- possibly some other helpers

### Run

- servers tests
  - specifically check setting of user and group as well as inheriting of the port
- services tests
- spec tests
- run tests
- integration tests
- identify server circular extending and error instead of current stack panic
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
