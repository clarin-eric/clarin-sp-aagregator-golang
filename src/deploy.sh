#!/bin/sh

set -e

#   /Users/wilelb/Code/work/clarin/git/infrastructure2/golang/src/clarin/shib-test
P="/Users/wilelb/Code/work/clarin/git/infrastructure2/golang"
BINARY="server"

echo "Building binary" && \
GOPATH=${P} CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o "${BINARY}_linux" *.go && \
echo "Building new docker image" && \
cp "${P}/src/clarin/shib-test/${BINARY}_linux" "/Users/wilelb/Code/work/clarin/git/infrastructure2/docker/docker-shib-test/image/" && \
cd /Users/wilelb/Code/work/clarin/git/infrastructure2/docker/docker-shib-test && \
sh build.sh --build --local && \
echo "Restarting docker container" && \
cd /Users/wilelb/Code/work/clarin/git/infrastructure2/docker/docker-clarin-nginx-proxy/run && \
docker-compose up -d --force-recreate shib-test && \
cd /Users/wilelb/Code/work/clarin/git/infrastructure2/golang/src/clarin/shib-test
