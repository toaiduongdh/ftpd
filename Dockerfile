
###################################
#Build stage

FROM golang:1.20-buster

RUN go install goftp.io/ftpd@3255cab36a

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        libssl-dev \
        git \
        bc \
        lsof \
        grep \
        kafkacat \
    && apt-get autoremove -y \
    && apt-get autoclean -y

