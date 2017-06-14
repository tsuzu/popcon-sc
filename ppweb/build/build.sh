#! /bin/bash

set -eu

export RDIR=$(dirname $0)
export DIR=$(cd $RDIR && pwd)
cd $DIR
rm ./ppweb 2> /dev/null || :
GOOS=linux GOARCH=amd64 go get -v -d ../...
GOOS=linux GOARCH=amd64 go build -o ./ppweb ..
docker build -t ppweb -f ./Dockerfile ..
