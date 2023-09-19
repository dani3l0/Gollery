CGO_ENABLED=0

deps:
	go get

build:
	go build -o dist/gollery

build-all:
	env GOOS=linux GOARCH=amd64 go build -o dist/gollery-amd64
	env GOOS=linux GOARCH=arm64 go build -o dist/gollery-arm64
	env GOOS=linux GOARCH=386 go build -o dist/gollery-386
	env GOOS=linux GOARCH=arm go build -o dist/gollery-arm
