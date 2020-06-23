FROM golang:1.11.2

ENV GO111MODULE=off

RUN go get k8s.io/code-generator; exit 0
RUN go get k8s.io/apimachinery; exit 0

RUN mkdir -p $repo

WORKDIR $GOPATH/src/k8s.io/code-generator
