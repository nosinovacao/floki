FROM golang:alpine AS builder
ENV GO111MODULE=on
# make g++ zlib-dev libssl1.1 libsasl zstd-libsapk add --no-cache  \
RUN apk add --update --no-cache \
                     bash              \
                     build-base        \
                     coreutils         \
                     gcc               \
                     git               \
                     g++               \
                     make              \
                     musl-dev          \
                     openssl-dev       \
                     openssl           \
                     libsasl           \
                     libgss-dev        \
                     rpm               \
                     lz4-dev           \
                     zlib-dev          \
                     ca-certificates

WORKDIR $GOPATH/src/floki/build/
COPY . .
RUN go mod download
RUN GOOS=linux GOARCH=amd64 go build -a -tags static_all -o /go/bin/floki

#FROM scratch
#COPY --from=builder /go/bin/floki /go/bin/floki
ENTRYPOINT ["/go/bin/floki"]
