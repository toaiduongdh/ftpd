
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


RUN go install github.com/toaiduongdh/ftpd@d88eeb6ce51347ecc6f5dcf3dc0c3d15827a5815

