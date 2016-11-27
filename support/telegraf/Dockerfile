FROM telegraf:1.1-alpine

RUN apk add --no-cache ca-certificates && \
    update-ca-certificates

COPY telegraf.conf /etc/telegraf/