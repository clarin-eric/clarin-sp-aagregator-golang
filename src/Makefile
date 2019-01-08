all: compile

compile:
	go get
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o "aagregator_linux" *.go
