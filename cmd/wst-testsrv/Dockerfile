# Image for running wst-testsrv in the Docker and Kubernetes integration tests.
#
# The build context is expected to contain a statically linked wst-testsrv binary built with:
#
#   CGO_ENABLED=0 go build -o wst-testsrv ./cmd/wst-testsrv
#
# busybox is used instead of scratch so that the image has `sh` and `kill` which are needed for executing commands
# and sending signals to the server in container environments.
FROM busybox:1.37.0-musl

COPY wst-testsrv /usr/local/bin/wst-testsrv

RUN mkdir -p /etc /var/www /run
