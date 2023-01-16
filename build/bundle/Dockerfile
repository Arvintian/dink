FROM ubuntu:22.04

RUN apt update && apt install -y runc fuse-overlayfs

COPY dist/ /app/
WORKDIR /app