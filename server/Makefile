export GOPATH := $(shell pwd)
export PATH := $(GOPATH)/bin:${PATH}
export GOBIN := $(GOPATH)/bin
.PHONY: all test clean

all: server

deps:
	go get -u github.com/abbot/go-http-auth
	go get -d github.com/libgit2/git2go
	cd src/github.com/libgit2/git2go && git checkout next && git submodule update --init &&	make install &&	go install
	cd ../../../..

server: bindata.go deps
	go build

bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

bindata.go: bin/go-bindata
	./bin/go-bindata -nometadata _static/...

clean:
	rm -rf bin
	rm -rf pkg
	rm -rf src
	rm -rf server
	rm -rf bindata.go

test:
	python test.py
