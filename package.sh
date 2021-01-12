#!/bin/bash
# 将文件打包进二进制包中方法
# 1. 安装go-bindata: go get -u github.com/jteeuwen/go-bindata/...@master
# 2. 在工程根目录下执行: go-bindata -o externalfile/externalfile.go -ignore=externalfile.go -pkg=externalfile externalfile/...

version="mysql-agent 1.0.0"
flags="-X 'main.version=$version' -X 'main.goVersion=`go version`' -X 'main.buildTime=`date +"%Y-%m-%d %H:%M:%S"`'"

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "$flags" -o mysql-agent agent/agent.go agent/version.go