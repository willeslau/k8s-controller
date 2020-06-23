FROM golang:1.14.4

RUN go get k8s.io/code-generator; exit 0
RUN go get k8s.io/apimachinery; exit 0

RUN mkdir -p ${GOPATH}/src/github.com/willeslau/k8s-controller

WORKDIR ${GOPATH}/src/github.com/willeslau/k8s-controller
