.PHONY: all build gotool clean help

BINARY="main"
PATH="build"
GO_VERSION=$(shell go version)
BUILD_TIME=$(shell date +%F-%Z/%T)
COMMIT_ID=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X 'main.GoVersion=${GO_VERSION}' -X main.BuildTime=${BUILD_TIME} -X main.CommitID=${COMMIT_ID}"

GO=$$GOROOT/build/go

all: gotool build

build:
	${GO} build ${LDFLAGS} -o ${PATH}/${BINARY}

gotool:
	${GO} fmt ./
	${GO} vet ./

clean:
	@if [ -f ${PATH}/${BINARY} ] ; then rm ${PATH}/${BINARY} ; fi

help:
	@echo "make - 格式化 Go 代码, 并编译生成二进制文件"
	@echo "make build - 编译 Go 代码, 生成二进制文件"
	@echo "make clean - 移除二进制文件和 vim swap files"
	@echo "make gotool - 运行 Go 工具 'fmt' and 'vet'"