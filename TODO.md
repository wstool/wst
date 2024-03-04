# TODO list

This should be a temporary just for the initial development to help quickly organise TODO items. Issues should be used
in the future.

## Code

### App
- possibly some other helpers

### Run
- replace chan error in Output and use instead some combined reader that returns immediately error
- local environment integration
- docker environment integration
- kubernetes environment integration
- dry run mode - integrate to action, services and sandboxes
- template rendering
- load testing action (wrk runner)
- tests

### Config
- local env port should be start / end instead of from / to maybe
- add checks for casting to lower integers (e.g. int16)
- merging - generic merging rules taken from params
- factories