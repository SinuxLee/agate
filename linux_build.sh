#!/usr/bin/env bash

set -e

export GOOS=linux
export GO111MODULE=on
export CGO_ENABLED=0
export GOARCH=amd64
branch=$(git branch | sed -n -e 's/^\* \(.*\)/\1/p')
datetime=$(date +%Y-%m-%d-%H-%M)
commit_id=$(git rev-list -n 1 HEAD | cut -c 1-7)
go_version=$(go version | cut -c 12-20)
project=svr
platform=$(uname)

if go build -ldflags "-X main.version=${branch}_${commit_id}_${datetime}_${go_version}" -mod=mod -o $project cmd/$project/main.go ; then
    if [ "$platform" == "Darwin" ] || echo "$platform" | grep -q "MINGW"; then
	      tar -czf $project.tar.gz $project
        rm $project
    else
        if [ ! -d "/tmp/deploy/ffa" ];then mkdir -p /tmp/deploy/ffa;fi
        mv $project /tmp/deploy/ffa/
    fi
fi