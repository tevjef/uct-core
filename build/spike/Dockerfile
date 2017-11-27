FROM alpine

WORKDIR /uct
RUN apk --no-cache add ca-certificates && update-ca-certificates
ENV PATH /uct:$PATH
COPY config.toml /uct
COPY spike /uct
EXPOSE 9876
EXPOSE 13000
CMD ["spike", "-c" , "/uct/config.toml"]