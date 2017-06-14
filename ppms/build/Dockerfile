FROM golang

MAINTAINER Tsuzu

RUN mkdir -p $GOPATH/src/github.com/cs3238-tsuzu/ && cd $GOPATH/src/github.com/cs3238-tsuzu/ && git clone -b swarm https://github.com/cs3238-tsuzu/popcon-sc 
COPY ./ $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppms
COPY ./build/ppms $GOPATH/bin/ppms

WORKDIR $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppms
ENTRYPOINT ["ppms"]
