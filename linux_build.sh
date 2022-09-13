#!/usr/bin/env bash

set -e

export GOOS=linux
export GO111MODULE=on
export CGO_ENABLED=0
export GOARCH=amd64
branch=$(git rev-parse --abbrev-ref HEAD)
datetime=$(date +%Y-%m-%d/%H:%M:%S)
commit_id=$(git rev-parse --short HEAD)
go_version=$(go version | awk '{print $3}')
ver_info="_branch:"${branch}_"commitid:"${commit_id}_"date:"${datetime}_"goversion:"${go_version}
project=svr
platform=$(uname)

if go build -ldflags "-X main.version=${ver_info}"  -mod=mod -o $project cmd/$project/*; then
    if [ "$platform" == "Darwin" ] || echo "$platform" | grep -q "MINGW"; then
	      tar -czf $project.tar.gz $project
        rm $project
    else
        if [ ! -d "/tmp/deploy/ffa" ];then mkdir -p /tmp/deploy/ffa;fi
        mv $project /tmp/deploy/ffa/
    fi
else
	  if echo "$platform" | grep -q "MINGW"; then
	      read -n 1
    fi
fi