#!/bin/sh

set -eu

go get -u github.com/tools/godep
godep save ./...
