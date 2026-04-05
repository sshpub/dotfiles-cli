VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS  := -s -w -X github.com/sshpub/dotfiles-cli/cmd.Version=$(VERSION)
BINARY   := dotfiles

.PHONY: build build-all test clean

build:
	go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY) .

build-all:
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 .
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 .
	GOOS=linux   GOARCH=386   go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-386 .
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe .
	GOOS=windows GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-arm64.exe .
	GOOS=windows GOARCH=386   go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-386.exe .
	GOOS=freebsd GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-freebsd-amd64 .

test:
	go test ./...

clean:
	rm -rf dist/
