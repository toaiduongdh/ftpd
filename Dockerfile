
###################################
#Build stage

FROM golang:1.20-buster

RUN go install github.com/toaiduongdh/ftpd@962762656db45053192fc66af4c6d7b2610a5c01

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
COPY config.sample.ini /config/config.ini

