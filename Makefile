prepare:
	- git clone git://github.com/libgit2/git2go.git $(GOPATH)/src/github.com/libgit2/git2go
	- cd $(GOPATH)/src/github.com/libgit2/git2go
	- git checkout origin/make-static
	- git submodule update --init
	- ./script/with-static.sh go install

build:
	go build -o pear

test:
	go test
