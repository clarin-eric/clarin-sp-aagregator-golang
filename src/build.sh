#!/bin/sh

set -ex

#   /Users/wilelb/Code/work/clarin/git/infrastructure2/golang/src/clarin/shib-aagregator
P="/Users/wilelb/Code/work/clarin/git/infrastructure2/golang"
BINARY="server"

GOPATH=${P} CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o "${BINARY}_linux" *.go && \
GOPATH=${P} CGO_ENABLED=0 GOARCH=amd64 GOOS=darwin go build -a -installsuffix cgo -ldflags="-s -w" -o "${BINARY}_osx" *.go && \
cp "${P}/src/clarin/shib-aagregator/${BINARY}_linux" "/Users/wilelb/Code/work/clarin/git/infrastructure2/docker/docker-shib-aagregator/image/"
