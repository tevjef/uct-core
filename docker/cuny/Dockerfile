FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY cuny /uct
COPY jet /uct

RUN apk --no-cache add ca-certificates && update-ca-certificates

EXPOSE 13304

CMD ["jet", "--input-format", "protobuf", "--output-format", "protobuf", "-c", "/uct/config.toml", "--scraper", "cuny", "-c", "/uct/config.toml"]
