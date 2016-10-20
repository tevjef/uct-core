FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY spike /uct

EXPOSE 9876

EXPOSE 13400

CMD ["spike", "-c" , "/uct/config.toml"]