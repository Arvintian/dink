FROM alpine:3.17

RUN apk update && apk add fuse-overlayfs runc nfs-utils

COPY dist/ /app/
WORKDIR /app