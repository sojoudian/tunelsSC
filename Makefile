.PHONY: default server client deps fmt clean all release-all test

BUILDTAGS=debug
default: all

deps:
	go mod download
	go mod verify

server: deps
	go build -tags '$(BUILDTAGS)' -o bin/ngrokd ./src/ngrok/main/ngrokd

fmt:
	go fmt ./...

client: deps
	go build -tags '$(BUILDTAGS)' -o bin/ngrok ./src/ngrok/main/ngrok

release-client: BUILDTAGS=release
release-client: client

release-server: BUILDTAGS=release
release-server: server

release-all: fmt release-client release-server

all: fmt client server

test:
	go test -v ./...

clean:
	go clean -i ./...
	rm -rf bin/

contributors:
	echo "Contributors to ngrok, both large and small:\n" > CONTRIBUTORS
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' | cut -d '<' -f1 >> CONTRIBUTORS