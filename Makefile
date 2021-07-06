.PHONY: build clean

VERSION=3.0.0
BUILD_TIME="$(shell date +%Y%m%d)"

GOPROXY=https://goproxy.cn,direct
GO_ENV=CGO_ENABLED=0 GOPROXY=$(GOPROXY) GOPRIVATE="repo.wooramel.cn"
GO_FLAGS=-ldflags="-X main.Version=$(VERSION) -X 'main.BuildTime=$(BUILD_TIME)' -extldflags -static"

all: build

build: clean
	@go build $(GO_FLAGS) -o ssr-subscriber ./cmd

test: clean
	@go test ./...

clean:
	@go clean ./...
	@rm -f ./ssr-subscriber
