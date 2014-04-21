set -x

set PROJECT_DIR=$(pwd)
git clone git://github.com/libgit2/git2go.git $GOPATH/src/github.com/git2go

chmod +x $GOPATH/src/github.com/libgit2/git2go/script/build-libgit2.sh

cd $GOPATH/src/github.com/libgit2/git2go
script/build-libgit2.sh

set LIBGIT2_LIB=$GOPATH/src/libgit2/git2go/libgit2/install/lib
export LD_LIBRARY_PATH=$LIBGIT2_LIB
export DYLD_LIBRARY_PATH=$LIBGIT2_LIB
export PKG_CONFIG_PATH=$LIBGIT2_LIB/pkgconfig

go install github.com/libgit2/git2go

cd PROJECT_DIR
