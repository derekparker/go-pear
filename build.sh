#!/bin/bash

mkdir -p builds
version=`git describe --abbrev=0 --tags`
echo $version
env GOOS=linux GOARCH=amd64 go build -o builds/pear-linux-$version
chmod +x builds/pear-linux-$version
echo "built linux version"
env GOOS=darwin GOARCH=amd64 go build -o builds/pear-macos-$version
chmod +x builds/pear-macos-$version
echo "built macos version"
env GOOS=windows GOARCH=amd64 go build -o builds/pear-windows-$version.exe
chmod +x builds/pear-windows-$version.exe
echo "built windows version"
