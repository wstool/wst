# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### Build, CI and Docs

- Update README and extend README - align with the current logic
- document parallel, not and reload actions in schema
- review the schema if it matches the config
- further extend README docs
- update Go to latest 1.23
- integration tests

### App

- organise Foundation - split to smaller pieces (especially the OS stuff)

### Run

- test and fix kubernetes environment
  - pods watching after deployment to identify that pod is running and catch CrashLoopBackOff and Error
- test and fix docker environment
  - container create fails if container already exist - remove the container like cli `docker container create --rm`
  - container wait does not finish even if the container is running - wait condition does not work
  - pulling of image is not awaited - waiting to fully download the image does not work
- extend and improve debug logging
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
- look into more consistent naming for public and private url vs local and private address
  - private has got different meaning in both
  - maybe address is not the best name
  - check what naming is used elsewhere and consider matching that
- look to removing Service Requires or rethink how it should work
  - if kept, it should define semantic what started really is (e.g. after checking start logs)
- consider moving server port to sandbox port
  - currently the server port is really just container specific and not used for local
  - consider more consistent naming differentiating that service port is public and server port is private
- kubernetes environment improvements
  - allow setting default kubeconfig to ~/.kube/config (will require home dir support)
  - support and test native hook start where command is nil (set better executable - ideally configurable)
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
- separate workspace for each environment and reset only the env that is being run
  - it's to keep the local for potential debugging
  - also move local env files under a single dir (compare to multiple _env dirs) and get rid of duplicated service naming in path
- add generation of execution shell script to easily start services in workspace
  - should be probably bin for each service in local env
  - consider what to do for Docker - maybe some simple Docker compose
  - consider what to do for Kubernetes - maybe yaml files and shell to start and stop them
- root mode execution
  - add support for template condition whether service starts under root (e.g. in containers)
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
