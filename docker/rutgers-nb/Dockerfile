FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY rutgers /uct
COPY jet /uct

RUN apk --no-cache add ca-certificates && update-ca-certificates

EXPOSE 13303

CMD ["jet", "--input-format", "protobuf", "--output-format", "protobuf", "--daemon", "1m", "-c", "/uct/config.toml", "--scraper-name", "rutgers-nb", "--scraper", "rutgers", "-c", "/uct/config.toml", "--campus", "NB", "--latest"]
