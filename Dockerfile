
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


RUN go install github.com/toaiduongdh/ftpd@cce8c9c3e914de159eb7e1491d2d57ec49d30b69
RUN go install github.com/toaiduongdh/ftpd/tools@cce8c9c3e914de159eb7e1491d2d57ec49d30b69

