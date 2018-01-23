FROM centos:7
RUN yum install git -y
#CERTS
RUN yum install wget ca-certificates -y
RUN update-ca-trust force-enable
RUN wget https://github.build.ge.com/CloudPod/ge_certificates/blob/master/GE_External_Root_CA_2.1.cer?raw=true -P /etc/pki/ca-trust/source/anchors/
RUN update-ca-trust extract

RUN wget https://dl.google.com/go/go1.9.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.9.3.linux-amd64.tar.gz
ENV PATH="$PATH:/usr/local/go/bin"
RUN mkdir -p /go/src
ENV GOPATH=/go
RUN yum install which -y
WORKDIR /root
ADD pre-reqs.sh .
RUN chmod +x pre-reqs.sh
RUN /bin/sh ./pre-reqs.sh
RUN mkdir -p /go/src/github.build.ge.com/212601587/



