LIBGIT2_LIB=$(GOPATH)/src/github.com/libgit2/git2go/libgit2/install/lib
PROJECT_DIR=$(pwd)
export LD_LIBRARY_PATH=$(LIBGIT2_LIB)
export DYLD_LIBRARY_PATH=$(LIBGIT2_LIB)
export PKG_CONFIG_PATH=$(LIBGIT2_LIB)/pkgconfig

prepare:
	- git clone git://github.com/libgit2/git2go.git $(GOPATH)/src/github.com/libgit2/git2go
	chmod +x $(GOPATH)/src/github.com/libgit2/git2go/script/build-libgit2.sh
	- cd $(GOPATH)/src/github.com/libgit2/git2go && script/build-libgit2.sh
	- go install github.com/libgit2/git2go

build:
	go build -o pear

test:
	go test
