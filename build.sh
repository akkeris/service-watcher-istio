#!/bin/sh

go get "github.com/stackimpact/stackimpact-go"
go get "github.com/akkeris/vault-client"
go get k8s.io/client-go/...
go get github.com/tools/godep
cd /go/src/k8s.io/client-go/
git checkout v8.0.0
godep restore ./...
cd /go/src/service-watcher-istio/
go build process.go

