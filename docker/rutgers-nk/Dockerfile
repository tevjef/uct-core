FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

EXPOSE 13301

ENTRYPOINT ["rutgers"]

CMD ["--campus", "NK", "--daemon", "2m", "--latest", "-f" , "protobuf", "-c" , "/uct/config.toml"]