export GOPATH := $(shell pwd)
export PATH := $(GOPATH)/bin:${PATH}
export GOBIN := $(GOPATH)/bin

.PHONY: all test clean cleanall deps
.INTERMEDIATE: bindata.go _static/version

TARGET=strapdown-server
RELEASE=no


PYTHON2=$(shell command -v python2 || command -v python)

all: $(TARGET)

deps:
	go get -u github.com/abbot/go-http-auth
	go get -d github.com/libgit2/git2go
	cd src/github.com/libgit2/git2go && git checkout next && git submodule update --init &&	make install &&	go install
	cd ../../../..

# PHONY target should not be a dependency
$(TARGET): bindata.go
	go build -o $@

bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

bindata.go: bin/go-bindata _static/version
	./bin/go-bindata -nometadata _static/...

_static/version:
ifeq (yes,$(RELEASE))
	$(PYTHON2) genversion.py build > $@
else
	$(PYTHON2) genversion.py > $@
endif

clean:
	rm -rf $(TARGET)
	rm -rf bindata.go
	rm -rf _static/version

cleanall: clean
	rm -rf bin
	rm -rf pkg
	rm -rf src

test:
	python test.py
