FROM golang:1.8.3-alpine3.5
# Install redis
RUN echo "http://dl-4.alpinelinux.org/alpine/v3.1/main" >> /etc/apk/repositories && \
    apk add --update redis=2.8.23-r0 && \
    rm -rf /var/cache/apk/* && \
    mkdir /data && \
    chown -R redis:redis /data && \
    echo -e "include /etc/redis-local.conf\n" >> /etc/redis.conf
RUN apk add --update git
WORKDIR /root
ADD pre-reqs.sh .
RUN chmod +x pre-reqs.sh
RUN /bin/sh ./pre-reqs.sh
WORKDIR /go/src/
