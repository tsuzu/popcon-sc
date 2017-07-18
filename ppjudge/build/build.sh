#! /bin/bash

if [[ -z "${TRAVIS_BRANCH}" ]]; then
    export BRANCH=$(git rev-parse --abbrev-ref HEAD)
else
    export BRANCH=$(if [ "$TRAVIS_PULL_REQUEST" == "false" ]; then echo $TRAVIS_BRANCH; else echo $TRAVIS_PULL_REQUEST_BRANCH; fi)
fi
set -eu

export RDIR=$(dirname $0)
export DIR=$(cd $RDIR && pwd)
cd $DIR
rm ./ppjudge 2> /dev/null || :
GOOS=linux GOARCH=amd64 go get -v -d ../...
GOOS=linux GOARCH=amd64 go build -o ./ppjudge ..
docker build --build-arg GIT_BRANCH=$BRANCH -t ppjudge -f ./Dockerfile ..
