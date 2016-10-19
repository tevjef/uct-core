FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

ENTRYPOINT ["ein"]
EXPOSE 13100

CMD ["-c" , "/uct/config.toml", "-f", "protobuf"]