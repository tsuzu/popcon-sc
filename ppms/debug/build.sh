#! /bin/bash

export RDIR=$(dirname $0)
export DIR=$(cd $RDIR && pwd)
cd $DIR
GOOS=linux GOARCH=amd64 go get -v -d ../...
GOOS=linux GOARCH=amd64 go build -o ./ppms ..
docker build -t ppms -f ./Dockerfile ..
