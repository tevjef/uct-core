FROM alpine

WORKDIR /uct
ENV PATH /uct:$PATH
COPY config.toml /uct
COPY edward /uct
RUN apk --no-cache add ca-certificates && update-ca-certificates
EXPOSE 2058
EXPOSE 13000
ENV GODEBUG "http2debug=2"
CMD ["edward", "-c" , "/uct/config.toml"]