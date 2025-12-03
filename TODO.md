# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### Build and CI

- update Go to the latest 1.25
- update deps
- update and re-run mockery
- integration tests

### Docs

#### README

- once sphinx docs available, clean up to only list the most important parts with links to docs

#### Sphinx

- create base structure for RST Sphinx docs
- take architecture from README and align it with the current logic
- document parameters handling and inheritance
- integrate some of the text from the blog
- create nicer config docs
- document template handling and functions
- come with some nice get started section

#### Schema

- document encode_path request action option
- document reload action in schema
- review the schema if it matches the config

#### Website

- look into setting up the website for Sphinx docs

### App

- move Kubernetes and Docker client to application part
  - it is so 100% test coverage can be required for run/
- organise Foundation - split to smaller pieces (especially the OS stuff)

### Run

#### Structure - Instances, Actions, Servers, Services
- fix request body chunk delay (with encoding none and low chunk size) that is not causing delay
- add template support to request body
- add typed parameters substitution for integers
  - this is to be able to, for example, parameterize status code
  - alternatively, it might be easier to allow automatic string to int conversion
  - this is support testing ProxyMatch in the basic base test
- extend metrics to allow requesting metrics in time
  - effectively, metrics should be stored in time series
  - when requested without time, it should define some operation to use for getting the result
    - should be per metric default - for example counter would be max, but elsewhere avg or other might make more sense
  - it would make sense to also support ranges
- bench action restructuring allowing different types of bench marking / load testing
  - move current parameters under attack
  - create a new type wave to do something similar like https://github.com/bjaspan/goofy
    - use time series of metrics
- extend metrics expectations to support checking metric in time
- support metrics server expectation
- custom server actions for parallel and not action
  - this is mainly for completeness with sequential and might be also useful in some cases
- support TLS config in bench action
  - this will likely require using custom transport
- extend TLS config for request and bench to support client cert
- support `protocols` field in bench action
  - extract the common logic
- extend protocols to support http3 in bench and request action
- extend request body to support multipart form data
  - it should support form fields
  - it should support files
- look into extending bench to support request body (see if all request action options will be possible)
- integrate better instance action identification
  - it should introduce name for each action and also pass parent name to nested actions in `parallel` or `not`
- add execute action custom environment variables support
  - should be an action map parameter
- look into output command matching for messages that are not found and waiting for timeout
  - basically if the message is not found, it waits for context timeout and not finding out that the output collector is done
  - it might need some way to unblock the scanner when there is nothing more to scan
- implement extended output extraction and validation
  - Integrate CEL expression as in https://gist.github.com/bukka/82d63d6144f4e82a3032517665af374f
  - Or alternatively there could be YAML like syntax https://gist.github.com/bukka/6891c6d2b59ddb8a89fd991ce658695f but that's possibly too custom
- look into default action service integration
  - it should be basically service defined in parent (e.g. sequential service) and used if no service is defined for the action
  - it could be then used in the string form like `expect//name`
  - this would be mainly useful for server actions where naming service currently creates dependency on name from action (that should not be required) 
- save UDS socket to /tmp if longer than 108 which might happen if workspace path is too long
  - ideally use format like this `/tmp/wst/{service_name}/{socket_name}.sock`
  - it should still allow too long UDS if socket_name is too long so this case can be tested
- look into more consistent naming for public and private url vs local and private address
  - private has got different meaning in both
  - maybe address is not the best name
  - check what naming is used elsewhere and consider matching that
- look into removing Service Requires or rethink how it should work
  - if kept, it should define semantic what started really is (e.g. after checking start logs)
- look into supporting multiple endpoints for service
  - for example, when multiple nginx servers are defined (one for http and one for https running on different ports)
  - this should be also somehow selected from request and other actions
- consider moving server port to sandbox port
  - currently the server port is really just container specific and not used for local
  - consider more consistent naming differentiating that service port is public and server port is private
- add support for ephemeral port allocation that should be the default if not ports specified
  - it should be also possible to overwrite port to ephemeral selection even if specified
- Add Temp dir support to Dirs as it might be useful, for example, for nginx temp paths
- consider adding support default Dirs so it is not required to specify in the config
  - could be either the actual enum name or another tag
- add a special resource file structure that could be used for scripts but also for certs and keys
  - this will also allow using custom paths and mode for certs and keys
- add support for more generating self-signed certificate
  - to allow automatic creation of certificate
  - possibly also look to what other certificate types could be useful
- enhance parameters merging
  - currently it's only one level (key on the first level overwrites everything) - consider recursive deep merging
- come up with custom error wrapping and types
  - eliminating differentiation based on error message for context deadline action check (e.g. in output action)
  - removal of deprecated (archived) github.com/pkg/errors
- stop action awaiting by checking if the task still runs
  - this might be useful for cases when it cannot be found from the logs
  - schema description in brach action-stop-await
- fuzzing action
- replace environment.ServiceSettings struct with environment.Service interface
  - code clean up really with saving some calls - it just messy to pre-create this struct
  - look into cleaner handling wp and env paths - use some struct with wp and env path rather than having 2 string maps

#### Execution

- add labels based filtering for the run
  - new options for selecting labels - only instances with those labels will run
- test dry run and how it works in all environments
- consider some internal options in the config
  - option to keep the old workspace rather than deleting - e.g. moving the whole dir to some archive - for debugging
- separate workspace for each environment and reset only the env that is being run
  - it's to keep the local for potential debugging
  - also move local env files under a single dir (compare to multiple _env dirs) and get rid of duplicated service naming in path
- add generation of execution shell script to easily start services in workspace
  - should be probably bin for each service in local env
  - consider what to do for Docker - maybe some simple Docker compose
  - consider what to do for Kubernetes - maybe yaml files and shell to start and stop them
- root mode execution
  - add support for template condition whether service starts under root (e.g. in containers)

#### Monitoring and debugging

- extend and improve debug logging
  - pattern matching does not need to repeat pattern for each match
    - or maybe put line first
  - there should be also log for successful debug log
  - log 'Task x started for service ...' rather than command
- test non debug logs - whether it is useful info and how errors are reported
- implement events to allow run event collection and possibly visualization of runs

#### Local environment

- look into issues with orphaned children which happen when a task main process dies suddenly
  - find a way to get them somehow killed so it doesn't require manual clean up/
- consider reporting closing output streams in Destroy
- find some smarter way for ports ranges so it does not need to be in each instance
  - maybe some global ports pool


#### Kubernetes environment

- pods watching after deployment to identify that pod is running and catch CrashLoopBackOff and Error
- support exec
- implement storing and handling certificates
  - this should go to secrets from rendered certificates
- allow setting default kubeconfig to ~/.kube/config (will require home dir support)
- support and test native hook start where command is nil (set better executable - ideally configurable)
- add health probes setup

#### Docker environment

- container create fails if container already exist - remove the container like cli `docker container create --rm`
- container wait does not finish even if the container is running - wait condition does not work
- pulling of image is not awaited - waiting to fully download the image does not work
- support exec
- implement storing and handling certificates
  - this is currently done using bind so check if there is a better (more secure) way to do it for private keye
- health check - waiting for container to be able to serve the traffic
- custom docker


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
