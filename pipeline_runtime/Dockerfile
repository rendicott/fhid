FROM centos:7
ARG CONFIG_FILE
ADD fhid .
ADD $CONFIG_FILE config.json
RUN ./fhid -version