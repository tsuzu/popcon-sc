#! /bin/bash

set -eu

export BRANCH=$(git branch)
export RDIR=$(dirname $0)
export DIR=$(cd $RDIR && pwd)
cd $DIR
rm ./ppjudge 2> /dev/null || :
GOOS=linux GOARCH=amd64 go get -v -d ../...
GOOS=linux GOARCH=amd64 go build -o ./ppjudge ..
docker build --build-arg GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD) -t ppweb -f ./Dockerfile ..
