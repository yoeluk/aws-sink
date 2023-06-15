.PHONY: lint test vendor clean copy_src

default: lint test

lint:
	golangci-lint run -v

test:
	go test -v -cover ./...

build:
	go build -v -o aws-sink

vendor:
	go mod vendor

clean:
	rm -rf ./vendor

copy_src:
	mkdir -p go/src/github.com/yoeluk/aws-sink
	cp -r aws local log s3 signer .traefik.yml go.mod Makefile sink.go sink_test.go go/src/github.com/yoeluk/aws-sink/