FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY hermes /uct

RUN apk --no-cache add ca-certificates && update-ca-certificates

EXPOSE 13201

CMD ["julia", "-c" , "/uct/config.toml"]