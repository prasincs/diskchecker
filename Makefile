go ?= go
TOP ?= .



_builds:
	mkdir -p $(TOP)/_builds/{linux,osx}

install:
	$(go) install ./...

getdeps: *.go
	$(go) get -u ./..

.deps-linux:
	apt-get update && apt-get install upx
	touch .deps-linux

getdeps-linux: getdeps .deps-linux

.deps-osx:
	brew install upx
	touch .deps-osx

getdeps-osx: .deps-osx

build-linux: getdeps-linux
	GOOS=linux $(go) build -o _builds/linux/diskchecker .
	upx --brute _builds/linux/diskchecker

build-osx: getdeps-osx
	GOOS=darwin $(go) build -o _builds/osx/diskchecker .
	upx --brute _builds/osx/diskchecker

all: install build-linux build-osx

clean:
	$(go) clean -i ./...
	rm -rf $(TOP)/_builds/

.PHONY: all clean getdeps build-linux build-osx
