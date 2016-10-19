FROM alpine

WORKDIR /uct

ENV PATH /uct:$PATH

COPY config.toml /uct

COPY spike /uct
RUN chmod +x /uct/spike

EXPOSE 9876
EXPOSE 13400

ENTRYPOINT ["spike", "-c" , "/uct/config.toml"]