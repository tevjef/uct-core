FROM alpine

WORKDIR /uct

RUN apk add --update ca-certificates && \
    rm -rf /var/cache/apk/* /tmp/* && \
    update-ca-certificates

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY spike /uct

EXPOSE 9876

EXPOSE 13400

CMD ["spike", "-c" , "/uct/config.toml"]