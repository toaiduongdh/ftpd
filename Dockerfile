
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


RUN go install github.com/toaiduongdh/ftpd@c68d11a33b4d764c069180b547d78d6e4027f55b
RUN go install github.com/toaiduongdh/ftpd/tools@c68d11a33b4d764c069180b547d78d6e4027f55b

