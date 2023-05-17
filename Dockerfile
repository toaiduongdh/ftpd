
###################################
#Build stage

FROM golang:1.20-buster


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


RUN go install github.com/toaiduongdh/ftpd@f5f9d754185ee9008566833bda4e6502a4504caf
RUN go install github.com/toaiduongdh/ftpd/kafkatools@3a4d247261112f6832ab416ada3e34cab55c2e68

