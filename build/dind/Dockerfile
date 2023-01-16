FROM docker:20.10.22-dind

RUN apk update && apk add fuse-overlayfs
ADD build/dind/daemon.json /etc/docker/daemon.json