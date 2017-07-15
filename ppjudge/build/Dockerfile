FROM golang:latest

VOLUME "/tmp"
ENV PPJUDGE_CGROUP=/sys/fs/cgroup
ENV PPJUDGE_DOCKER=unix:///var/run/docker.sock
ENV PPJUDGE_ON_DOCKER=1
ENV PPJUDGE_DOCKER_VER=v1.29
ARG GIT_BRANCH=${GIT_BRANCH:-swarm}

RUN mkdir -p $GOPATH/src/github.com/cs3238-tsuzu/ && cd $GOPATH/src/github.com/cs3238-tsuzu/ && git clone -b $GIT_BRANCH https://github.com/cs3238-tsuzu/popcon-sc 
COPY ./ $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppjudge
COPY ./build/ppjudge $GOPATH/bin/ppjudge

WORKDIR $GOPATH/src/github.com/cs3238-tsuzu/popcon-sc/ppjudge
ENTRYPOINT ["ppjudge"]
