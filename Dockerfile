FROM golang:1.13-alpine

RUN apk update
RUN apk add openssl ca-certificates git curl
ENV GO111MODULE on
RUN mkdir -p /root/.kube/certs
WORKDIR /go/src/github.com/akkeris/node-watcher-f5
COPY . .
RUN go build .
ADD start.sh /start.sh
RUN chmod +x /start.sh
CMD "/start.sh"



