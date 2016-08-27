
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
exec voletc >> /var/log/$(NAME).log 2>&1
endef

define VOLETC_INSTALL
mkdir -p /usr/local/bin\n
mv ./voletc /usr/local/bin/\n
mkdir -p /etc/init/\n
mv ./voletc.conf /etc/init/
endef

export VOLETC_STARTUP
export VOLETC_INSTALL

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
	GOOS=linux CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w" -o ./build/linux/$(NAME) .

.darwin-build:
	if [ -e ./build/darwin ]; then rm -rf ./build/darwin; fi
	mkdir -p ./build/darwin

	GOOS=darwin CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w" -o ./build/darwin/$(NAME) .

.PHONY: install.sh
install.sh:
	cd ./build/linux && echo $${VOLETC_INSTALL} > install.sh
	chmod +x ./build/linux/install.sh

.PHONY: voletc.conf
voletc.conf:
	rm -rf ./build/linux
	mkdir -p ./build/linux
	cd ./build/linux && echo $${VOLETC_STARTUP} > voletc.conf

# Should be run after make all
.PHONY: installer
installer:
	sea ./build/linux/ $(NAME)-installer.sh voletc ./install.sh
	mv $(NAME)-installer.sh ./build/linux
	cd ./build/linux && tar -czvf $(NAME)-$(VERSION)-linux.tgz $(NAME)-installer.sh && mv $(NAME)-$(VERSION)-linux.tgz ../
	cd ./build/darwin && tar -czvf $(NAME)-$(VERSION)-darwin.tgz $(NAME) && mv $(NAME)-$(VERSION)-darwin.tgz ../

all: .darwin-build .linux-build install.sh

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
