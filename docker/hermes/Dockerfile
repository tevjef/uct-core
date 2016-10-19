FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

ENTRYPOINT ["hermes"]

EXPOSE 13200

CMD ["-c" , "/uct/config.toml", "-d"]