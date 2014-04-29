LIBGIT2_LIB=$(GOPATH)/src/github.com/libgit2/git2go/libgit2/install/lib
PROJECT_DIR=$(pwd)
export LD_LIBRARY_PATH=$(LIBGIT2_LIB)
export DYLD_LIBRARY_PATH=$(LIBGIT2_LIB)
export PKG_CONFIG_PATH=$(LIBGIT2_LIB)/pkgconfig
LIBGIT_SRC_PATH=$(GOPATH)/src/github.com/libgit2/git2go/libgit2
LIBGIT_INSTALL_PREFIX=$(LIBGIT_SRC_PATH)/install

git2go: libgit
	CGO_LDFLAGS="$(LIBGIT_INSTALL_PREFIX)/lib/libgit2.a `pkg-config --libs --static $(LIBGIT_INSTALL_PREFIX)/lib/pkgconfig/libgit2.pc`" \
	go install github.com/libgit2/git2go

libgit:
	mkdir -p $(LIBGIT_SRC_PATH)/build
	cd $(LIBGIT_SRC_PATH)/build && \
	cmake .. -DCMAKE_INSTALL_PREFIX=$(LIBGIT_INSTALL_PREFIX) \
	    -DTHREADSAFE=ON \
	    -DBUILD_CLAR=OFF \
	    -DBUILD_SHARED_LIBS=OFF \
	    -DCMAKE_C_FLAGS=-fPIC && \
	cd .. && \
	cmake --build . && \
	cmake --build . --target install

# prepare: git2go
	# - git clone git://github.com/libgit2/git2go.git $(GOPATH)/src/github.com/libgit2/git2go
	# chmod +x $(GOPATH)/src/github.com/libgit2/git2go/script/build-libgit2.sh
	# - cd $(GOPATH)/src/github.com/libgit2/git2go && script/build-libgit2.sh
	# - go install github.com/libgit2/git2go

build:
	go build -o pear

test:
	go test
