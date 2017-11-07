export GOPATH := $(shell pwd)
export PATH := $(GOPATH)/bin:${PATH}
export GOBIN := $(GOPATH)/bin

.PHONY: all test clean cleanall deps static
.INTERMEDIATE: bindata.go _static/version

TARGET=strapdown-server
RELEASE=no


PYTHON2=$(shell command -v python2 || command -v python)

all: $(TARGET)

deps:
	go get -u github.com/abbot/go-http-auth
	go get -u github.com/jteeuwen/go-bindata/...
	go get -u golang.org/x/oauth2/google
	go get -u google.golang.org/api/admin/directory/v1
	go get -d github.com/libgit2/git2go
	cd src/github.com/libgit2/git2go/ && git checkout next && git submodule update --init && cd vendor/libgit2 && mkdir -p install/lib && mkdir -p build && cd build && cmake -DBUILD_CLAR=OFF -DBUILD_SHARED_LIBS=OFF -DUSE_OPENSSL=OFF -DUSE_SSH=OFF -DCURL=OFF -DCMAKE_C_FLAGS=-fPIC -DCMAKE_BUILD_TYPE="RelWithDebInfo" -DCMAKE_INSTALL_PREFIX=../install .. && cmake --build . 
	cd src/github.com/libgit2/git2go && go install ./...

# PHONY target should not be a dependency
$(TARGET): bindata.go
	go build -ldflags='-s -w' $(if $(BUILD_STATIC),-ldflags='-extldflags "-static"') -o $@

static:
	grunt

bindata.go: _static/version
	bin/go-bindata -nometadata _static/...

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
	$(PYTHON2) test.py
