#! /bin/bash
set -eu
cd $(dirname $0)
docker build -t sendmail .
