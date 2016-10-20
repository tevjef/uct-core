FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY rutgers /uct

EXPOSE 13302

CMD ["rutgers", "--campus", "CM", "--daemon", "2m", "--latest", "-f" , "protobuf", "-c" , "/uct/config.toml"]