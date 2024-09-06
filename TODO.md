# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### Build, CI and Docs

- update Go to latest 1.22 (re-run mockery)
- update deps
- add pipeline to run unit test
- document parallel, not and reload actions in schema
- review the schema if it matches the config
- extend and update README docs
- update Go to latest 1.23
- integration tests

### App

- organise Foundation - split to smaller pieces (especially the OS stuff)

### Run

- test and fix kubernetes environment
- test and fix docker environment
- identify server circular extending and error instead of current stack panic
- identify template include recursion (nesting limit)
- currently nothing is printed if the start binary does not exist - it should print proper error
  - change fpm_binary to php-f to recreate
- extend and improve debug logging
  - bench action should log vegeta command alternative and a bit more info
  - request - sending request should not print request and response struct but properly format it for better readability
  - message "Creating config for paths" should not use %v for overwrites ideally 
  - storing response metrics logs some object reference
  - pattern matching does not need to repeat pattern for each match
  - or maybe put line first
  - there should be also log for successful debug log
  - log 'Task x started for service ...' rather than command
- test non debug logs - whether it is useful info and how errors are reported
- integrate better instance action identification
  - it should introduce name for each action and also pass parent name to nested actions in `parallel` or `not`
- introduce sequential action for more complex scenarios (e.g. seq task in parallel action)
  - might be worth to consider whether top action should be wrapped to reduce code needed
- custom server actions for sequential action
  - useful to wrap multiple action - e.g. fpm start + expectations
- support metrics server expectation
- look into doing some partial expectation
  - some sort of contains mode rather than full match
- add labels based filtering for the run
  - new options for selecting labels - only instances with those labels will run
- look to removing Service Requires or rethink how it should work
  - if kept, it should define semantic what started really is (e.g. after checking start logs)
- kubernetes environment improvements
  - add health probes setup
- docker environment improvements
  - health check - waiting for container to be able to serve the traffic
  - custom docker
- local environment improvements
  - support for UDS in address
  - consider reporting closing output streams in Destroy
- test dry run and how it works in all environments
- consider some internal options in the config
  - option to keep the old workspace rather than deleting - e.g. moving the whole dir to some archive - for debugging
- enhance parameters merging
  - currently it's only one level (key on the first level overwrites everything) - consider recursive deep merging
- add generation of execution shell script to easily start services in workspace
  - should be probably bin for each service
- root mode execution
- come up with custom error wrapping and types
  - eliminating differentiation based on error message for context deadline action check (e.g. in output action)
  - removal of deprecated (archived) github.com/pkg/errors
- replace environment.ServiceSettings struct with environment.Service interface
  - code clean up really with saving some calls - it just messy to pre-create this struct
- fuzzing action

### List

- Introduce list command to list all instances
  - it should also allow listing description
  - some basic search

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
