FROM centos:7

RUN yum install golang git -y
RUN yum install epel-release -y
RUN yum update -y
RUN yum install redis -y
RUN mkdir -p /go/src
ENV GOPATH=/go/src
RUN go get "golang.org/x/net/context"
WORKDIR /root
ADD pre-reqs.sh .
RUN chmod +x pre-reqs.sh
RUN /bin/sh ./pre-reqs.sh
WORKDIR /go/src/



