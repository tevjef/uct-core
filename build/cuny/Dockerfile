FROM alpine

WORKDIR /uct
ENV PATH /uct:$PATH
COPY config.toml /uct
COPY cuny /uct
COPY jet /uct
RUN apk --no-cache add ca-certificates && update-ca-certificates
EXPOSE 13000
CMD ["jet", "-c", "/uct/config.toml", "--scraper", "cuny", "-c", "/uct/config.toml"]
