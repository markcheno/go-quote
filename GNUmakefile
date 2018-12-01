#
#	Makefile for hookAPI
#
# switches:
#	define the ones you want in the CFLAGS definition...
#
#	TRACE		- turn on tracing/debugging code
#
#
#
#

# Version for distribution
VER=1_0r1
GOPATH=$(shell go env GOPATH):$(PWD)

export GOPATH
MAKEFILE=GNUmakefile

# We Use Compact Memory Model

all: bin/quote
	@[ -d bin ] || exit

bin/quote: quote/main.go
	@[ -d bin ] || mkdir bin
	(cd quote; go build -o ../bin/quote)
	@strip $@ || echo "quote OK"

win64: bin/quote64.exe
win32: bin/quote32.exe

bin/quote64.exe: bin quote/main.go
	(cd quote; GOOS=windows GOARCH=amd64 go build -o ../bin/quote64.exe)

bin/quote32.exe: bin quote/main.go
	(cd quote; GOOS=windows GOARCH=386 go build -o ../bin/quote32.exe)

clean:

distclean: clean
	@rm -rf bin
