
NAMESPACE = ipkg
NAME = voletc
VERSION = $(shell grep "const VERSION" version.go | cut -d "\"" -f 2)

PROJPATH = github.com/$(NAMESPACE)/$(NAME)

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

define VOLETC_STARTUP
description "voletc plugin"\n
start on (local-filesystems and net-device-up IFACE=eth0)\n
stop on runlevel [!12345]\n
exec voletc -server >> /var/log/$(NAME).log 2>&1
endef

export VOLETC_STARTUP

clean:
	rm -rf ./build
	rm -f ./coverage.out
	rm -rf ./testrun
	go clean -i ./...

.run-consul:
	docker run -d -p 127.0.0.1:8500:8500 --name consul progrium/consul -server -bootstrap 

.PHONY: test
test:
	go test -cover ./...

.PHONY: deps
deps:
	go get -d -v ./...

.linux-build: voletc.conf
	mkdir -p ./build/linux/usr/local/bin
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w" -o ./build/linux/usr/local/bin/$(NAME) .

.darwin-build:
	mkdir -p ./build/darwin
	GOOS=darwin CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w" -o ./build/darwin/$(NAME) .

.PHONY: voletc.conf
voletc.conf:
	mkdir -p ./build/linux/etc/init/
	echo $${VOLETC_STARTUP} > ./build/linux/etc/init/voletc.conf

# Should be run after make all
.PHONY: installer
installer:
	cd ./build && tar -czf $(NAME)-$(VERSION)-linux.tgz -C ./linux/ .
	cd ./build && tar -czf $(NAME)-$(VERSION)-darwin.tgz  -C ./darwin/ .

all: .darwin-build .linux-build

.docker-test:
	docker run --link consul:consul --rm -v $(shell pwd):/go/src/${PROJPATH} -w /go/src/${PROJPATH} golang:1.6.3 make clean deps test

.docker-build:
	docker run --rm -v $(shell pwd):/go/src/${PROJPATH} -w /go/src/${PROJPATH} golang:1.6.3 make clean deps all

# Assemble image
.docker-image:
	docker build --no-cache -t $(NAMESPACE)/$(NAME):$(VERSION) .

# Complete docker build
.PHONY: docker
docker: .docker-build .docker-image
