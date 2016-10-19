FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

EXPOSE 13303

ENTRYPOINT ["rutgers", "--campus", "NB", "--daemon", "2m", "--latest", "-f" , "protobuf", "-c" , "/uct/config.toml"]