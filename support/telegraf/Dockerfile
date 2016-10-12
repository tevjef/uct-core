FROM telegraf:alpine

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY telegraf.conf /etc/telegraf/