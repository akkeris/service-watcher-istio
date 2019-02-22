FROM golang:1.10

RUN mkdir -p /go/src/service-watcher-istio
ADD process.go  /go/src/service-watcher-istio/process.go
ADD services /go/src/service-watcher-istio/services
ADD k8sconfig /go/src/service-watcher-istio/k8sconfig
ADD utils /go/src/service-watcher-istio/utils

ADD build.sh /build.sh
RUN chmod +x /build.sh
RUN /build.sh

RUN mkdir -p /root/.kube/certs
ADD start.sh /start.sh
RUN chmod +x /start.sh
CMD "/start.sh"



