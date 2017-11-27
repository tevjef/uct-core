FROM golang:1.9-alpine AS build-env

COPY . /go/src/github.com/tevjef/uct-core

ENV PATH /go/bin:$PATH

RUN apk --no-cache add ca-certificates openssl git && \
    update-ca-certificates && \
    go get -u github.com/golang/dep/cmd/dep && \
    cd /go/src/github.com/tevjef/uct-core && \
    dep ensure -v

WORKDIR /go/src/github.com/tevjef/uct-core

RUN go install -ldflags "-s -w" -v github.com/tevjef/uct-core/jet
RUN go install -ldflags "-s -w" -v github.com/tevjef/uct-core/scrapers/cuny

FROM alpine
ENV PATH /:$PATH
WORKDIR /
RUN apk --no-cache add ca-certificates openssl && \
    update-ca-certificates
COPY --from=build-env /go/bin/jet /
COPY --from=build-env /go/bin/cuny /
COPY --from=build-env /go/src/github.com/tevjef/uct-core/common/conf/config.toml /
RUN apk --no-cache add ca-certificates && update-ca-certificates
EXPOSE 13000
CMD ["jet", "-c", "config.toml", "--scraper", "cuny", "-c", "config.toml"]