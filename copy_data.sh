#!/bin/bash

#_GOPATH="/Users/wilelb/Code/work/clarin/git/infrastructure2/golang"
_BINARY="aagregator"

init_data (){
    LOCAL=0
    if [ "$1" == "local" ]; then
        LOCAL=1
    fi

    if [ "${LOCAL}" -eq 0 ]; then
        #Remote / gitlab ci
	echo "Building ${_BINARY}_linuxi remotely"
        cd ..
        docker run --rm -v "$PWD/src":/go/src/clarin/shib-aagregator/src -w /go/src/clarin/shib-aagregator/src golang:1.8 make
        mv "src/${_BINARY}_linux" ./image
        cd image
    else
        cd ..
        echo "Building ${_BINARY}_linuxi locally"
#        GOPATH=${_GOPATH} CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o "${_BINARY}_linux" *.go && \
	docker run --rm -v "$PWD/src":/go/src/clarin/shib-aagregator/src -w /go/src/clarin/shib-aagregator/src golang:1.8 make
        mv "src/${_BINARY}_linux" ./image && \
        cd ./image
    fi
}

cleanup_data () {
    echo "Removing ${_BINARY}_linux" && \
    rm -f "${_BINARY}_linux"
}
