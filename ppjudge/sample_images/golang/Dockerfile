FROM golang:1.8

LABEL ppjudge.language=golang
LABEL ppjudge.language.version=1.8.0
LABEL ppjudge.os=ubuntu-16.04
ENV PATH $PATH:/usr/local/go/bin:/go/bin

RUN apt-get update && apt-get upgrade -y && apt-get install -y time sudo
#RUN echo "Defaults env_keep += \"PATH\"" >> /etc/sudoers
#RUN echo "Defaults !env_reset" >> /etc/sudoers
#RUN echo "Defaults secure_path += \":/usr/local/go/bin:/go/bin\"" >> /etc/sudoers

RUN echo "Defaults	env_reset" >> /etc/sudoers
RUN echo "Defaults	secure_path=\"/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/go/bin\"" >> /etc/sudoers
