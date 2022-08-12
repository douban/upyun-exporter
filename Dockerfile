FROM golang:alpine as build-env

RUN apk add git

COPY . /go/src/github.com/douban/upyun-exporter
WORKDIR /go/src/github.com/douban/upyun-exporter
# Build
ENV GOPATH=/go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -v -a -ldflags "-s -w" -o /go/bin/upyun-exporter .

FROM library/alpine:3.15.0
COPY --from=build-env /go/bin/upyun-exporter /usr/bin/upyun-exporter
ENTRYPOINT ["upyun-exporter"]
