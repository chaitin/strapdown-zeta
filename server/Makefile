export GOPATH := $(shell pwd)
export PATH := $(GOPATH)/bin:${PATH}
export GOBIN := $(GOPATH)/bin
.PHONY: all static_file

all: server

deps:
	go get -u github.com/abbot/go-http-auth
	go get -d github.com/libgit2/git2go
	cd src/github.com/libgit2/git2go && git checkout next && git submodule update --init &&	make install &&	go install
	cd ../../../..

server: static_file deps
	go build

go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

static_file: go-bindata
	rm -rf go-bindata.go
	./bin/go-bindata -nometadata _static/...

clean:
	rm -rf bin
	rm -rf pkg
	rm -rf src
	rm -rf server

test:
	python test.py
