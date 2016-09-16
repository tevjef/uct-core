FROM telegraf:alpine

RUN apk update && \
    apk add ca-certificates && \
    update-ca-certificates

COPY telegraf.conf /etc/telegraf/