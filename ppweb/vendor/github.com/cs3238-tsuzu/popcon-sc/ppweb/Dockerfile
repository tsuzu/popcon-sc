FROM golang

MAINTAINER Tsuzu

RUN mkdir -p $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppweb
COPY . $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppweb
RUN go get github.com/cs3238-tsuzu/popcon-sc/ppweb

WORKDIR $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppweb
ENTRYPOINT ["ppweb"]
