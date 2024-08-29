# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### Build and CI

- update Go to latest 1.22 and then 1.23 (re-run mockery)
- update deps
- add pipeline to run unit test
- integration tests

### App

- organise Foundation - split to smaller pieces (especially the OS stuff)

### Run

- scripts are currently not created in their files but instead directly in /var/www which is a file
- fix port allocation - currently it is always 0 for services
- make Service Address hostname configurable so it's not always 0.0.0.0
  - consider support for UDS
- look to removing Service Requires or rethink how it should work
  - if kept, it should define semantic what started really is (e.g. after checking start logs)
- identify server circular extending and error instead of current stack panic
- identify template include recursion (nesting limit)
- add detailed info and debug logging
- kubernetes environment improvements
  - add health probes setup
- docker environment improvements
  - health check - waiting for container to be able to serve the traffic
  - custom docker 
- consider reporting closing output streams in localEnvironment Destroy
- support metrics server expectation

### Config

- parsing - if instance timeouts is not specified, the default action 30000 default is not applied
  - this might be a generic problem that nested struct parsing is skipped if not present
- parsing - add logic to use file name for instance name if present
- parsing - improve error messages
  - clean up and make error messages consistent
  - verify that location is correct and not missing leave
- parsing - warn on unknown fields in the config for struct mapping
  - for example setting environment ports `from` and `to` fields should notify that only `start` and `end` supported
- parsing - config version should allow number - not just string so 1.0 can be used instead of "1.0"
- parsing - support time and metric units (auto string to number conversions)
  - e.g. 1k = 100, 1s = 1000, 1ms = 1
- merging - generic merging rules taken from params
- boolean conversion in the same way as string (maybe something more generic that can handle both)
