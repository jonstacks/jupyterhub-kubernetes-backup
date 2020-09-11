FROM golang:alpine as builder
ENV CGO_ENABLED=0
RUN apk add --no-cache alpine-sdk

WORKDIR /app
COPY . .
RUN make install

FROM alpine
COPY --from=builder /go/bin/* /usr/local/bin/