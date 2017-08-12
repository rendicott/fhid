FROM centos:7

RUN yum install golang git -y
RUN yum install epel-release -y
RUN yum update -y
RUN yum install redis -y
RUN mkdir -p /go/src
ENV GOPATH=/go
WORKDIR /root
ADD pre-reqs.sh .
RUN chmod +x pre-reqs.sh
RUN /bin/sh ./pre-reqs.sh
RUN mkdir -p /go/src/github.build.ge.com/212601587/



