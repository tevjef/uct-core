FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY ein /uct

EXPOSE 13100

CMD ["ein", "-c" , "/uct/config.toml", "-f", "protobuf"]