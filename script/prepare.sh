set -x

go get github.com/libgit2/git2go

chmod +x $GOPATH/src/github.com/libgit2/git2go/script/build-libgit2.sh
$GOPATH/src/github.com/libgit2/git2go/script/build-libgit2.sh

export LD_LIBRARY_PATH=libgit2/install/lib
export DYLD_LIBRARY_PATH=libgit2/install/lib
export PKG_CONFIG_PATH=libgit2/install/lib/pkgconfig
