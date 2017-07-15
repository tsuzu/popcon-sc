FROM golang

MAINTAINER Tsuzu

ARG GIT_BRANCH=${GIT_BRANCH:-swarm}

RUN mkdir -p $GOPATH/src/github.com/cs3238-tsuzu/ && cd $GOPATH/src/github.com/cs3238-tsuzu/ && git clone -b $GIT_BRANCH https://github.com/cs3238-tsuzu/popcon-sc 
COPY ./ $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppjc
COPY ./build/ppjc $GOPATH/bin/ppjc

WORKDIR $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppjc
ENTRYPOINT ["ppjc"]
