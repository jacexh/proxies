FROM golang:1.21-alpine as builder
WORKDIR /go/src/
COPY . /go/src/
ENV GOPROXY https://goproxy.cn,direct
RUN go build -v -o proxies cmd/main.go

FROM alpine:3.8
WORKDIR /app
COPY --from=builder /go/src/proxies .
CMD ["/app/proxies"]