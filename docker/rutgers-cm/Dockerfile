FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY rutgers /uct

RUN apk --no-cache add ca-certificates && update-ca-certificates

EXPOSE 13302

CMD ["rutgers", "--campus", "CM", "--daemon", "2m", "--latest", "-f" , "protobuf", "-c" , "/uct/config.toml"]