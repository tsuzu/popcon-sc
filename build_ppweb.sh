#! /bin/bash

mkdir -p ppweb/debug
rm ./ppweb/debug/ppweb
export GOOS=linux
export GOARCH=amd64
cd ppweb && go build -o ./debug/ppweb

export PP_DB_PASSWORD="test"
export PP_TOKEN="test"
docker-compose build ppweb
