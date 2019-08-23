FROM golang:alpine AS builder
ENV GO111MODULE=on
RUN apk add --update --no-cache git librdkafka-dev pkgconf bash make g++ zlib-dev libssl1.1 libsasl zstd-libs && \
    git clone https://github.com/edenhill/librdkafka.git && \
    cd librdkafka && \
    ./configure --prefix /usr && \
    make -j8 && \
    make install

WORKDIR $GOPATH/src/logtail/build/
COPY . .
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -a -tags static_all -o /go/bin/logtail

FROM scratch
COPY --from=builder /go/bin/logtail /go/bin/logtail
ENTRYPOINT ["/go/bin/logtail"]
