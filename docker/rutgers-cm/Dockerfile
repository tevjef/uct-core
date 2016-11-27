FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY rutgers /uct
COPY jet /uct

RUN apk --no-cache add ca-certificates && update-ca-certificates

EXPOSE 13302

CMD ["jet", "--input-format", "protobuf", "--output-format", "protobuf", "--daemon", "1m", "-c", "/uct/config.toml", "--scraper-name", "rutgers-cm", "--scraper", "rutgers", "-c", "/uct/config.toml", "--campus", "CM", "--latest"]
