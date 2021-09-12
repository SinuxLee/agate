#!/usr/bin/env bash

set -e

export GOOS=linux
export GO111MODULE=on
export CGO_ENABLED=0
export GOARCH=amd64
project=svr
platform=$(uname)

if go build -mod=mod -o $project cmd/$project/main.go ; then
    if [ "$platform" == "Darwin" ] || echo "$platform" | grep -q "MINGW"; then
	   tar -czf $project.tar.gz $project
       rm $project
    else
        if [ ! -d "/tmp/deploy/ffa" ];then mkdir -p /tmp/deploy/ffa;fi
        mv $project /tmp/deploy/ffa/
    fi
fi