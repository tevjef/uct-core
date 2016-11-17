FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY rutgers-scraper /uct
COPY jet /uct

RUN apk --no-cache add ca-certificates && update-ca-certificates

EXPOSE 13301

CMD ["jet", "--input-format", "json", "--output-format", "protobuf", "--daemon", "5m", "-c", "/uct/config.toml", "--scraper-name", "rutgers-nk", "--scraper", "rutgers-scraper", "-c", "/uct/config.toml", "--campus", "NK"]
