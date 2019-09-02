FROM alpine

WORKDIR /uct
ENV PATH /uct:$PATH
COPY config.toml /uct
COPY edward /uct
EXPOSE 2058
EXPOSE 13000
CMD ["edward", "-c" , "/uct/config.toml"]