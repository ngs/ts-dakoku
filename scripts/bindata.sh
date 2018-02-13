#!/bin/sh

set -eu

go get -u github.com/jteeuwen/go-bindata/...
go-bindata -pkg app -o ./app/bindata.go assets
