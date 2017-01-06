FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY njit /uct
COPY jet /uct

RUN apk --no-cache add ca-certificates && update-ca-certificates

EXPOSE 13304

CMD ["jet", "-c", "/uct/config.toml", "--scraper-name", "njit", "--scraper", "njit", "-c", "/uct/config.toml"]

