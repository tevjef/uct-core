FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY hermes /uct

EXPOSE 13200

CMD ["hermes", "-c" , "/uct/config.toml", "-d"]