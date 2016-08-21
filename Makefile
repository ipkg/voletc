
NAMESPACE = ipkg
NAME = voletc
VERSION = $(shell grep "const VERSION" version.go | cut -d "\"" -f 2)

INST_PKG_NAME = $(NAME)-$(VERSION)-$(shell go env GOOS).tgz

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
	go clean -i ./...
	rm -f $(NAME)-installer

.PHONY: test
test: clean
	docker run -d -p 127.0.0.1:8500:8500 --name consul progrium/consul -server -bootstrap
	@sleep 3;
	go test -coverprofile=coverage.out ./...; docker rm -f consul


.PHONY: deps
deps:
	go get -d -v ./...

build: clean voletc.conf install.sh
	[ -d ./build ] || mkdir ./build
	CGO_ENABLED=0 go build -a -tags netgo -installsuffix netgo -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w" .
	mv ${NAME} ./build/

.docker-build:
	docker run --rm -v $(shell pwd):/go/src/${PROJPATH} -w /go/src/${PROJPATH} golang:1.6.3 make clean deps build

# Assemble image
.docker-image:
	docker build --no-cache -t $(NAMESPACE)/$(NAME):$(VERSION) .

# Complete docker build
.PHONY: docker
docker: .docker-build .docker-image

.PHONY: install.sh
install.sh:
	[ -d ./build ] || mkdir ./build
	cd ./build && echo $${VOLETC_INSTALL} > install.sh
	chmod +x ./build/install.sh

.PHONY: voletc.conf
voletc.conf:
	[ -d ./build ] || mkdir ./build
	cd ./build && echo $${VOLETC_STARTUP} > voletc.conf

.PHONY: installer
installer: build
	sea ./build/ $(NAME)-installer voletc ./install.sh 
	tar -czvf $(INST_PKG_NAME) $(NAME)-installer
	mv $(NAME)-installer ./build/
	mv $(INST_PKG_NAME) ./build/
	