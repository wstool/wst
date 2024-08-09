# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### App
- possibly some other helpers

### Run

- integration tests
- look to removing Service Requires or rethink how it should work
  - if kept, it should define semantic what started really is (e.g. after checking start logs)
- identify server circular extending and error instead of current stack panic
- add detailed info and debug logging
- kubernetes environment improvements
  - add health probes setup
- docker environment improvements
  - health check - waiting for container to be able to serve the traffic
  - custom docker 
- support metrics server expectation

### Config

- parsing - improve error messages
  - clean up and make error messages consistent
  - verify that location is correct and not missing leave
- parsing - config version should allow number - not just string so 1.0 can be used instead of "1.0"
- merging - generic merging rules taken from params
- boolean conversion in the same way as string (maybe something more generic that can handle both)
