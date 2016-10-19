FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

EXPOSE 13302

ENTRYPOINT ["rutgers", "--campus", "CM", "--daemon", "2m", "--latest", "-f" , "protobuf", "-c" , "/uct/config.toml"]