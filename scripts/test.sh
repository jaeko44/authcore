#!/bin/sh

set -e

BASEDIR=$( cd "$(dirname "$0")"/.. ; pwd -P )
cd "$BASEDIR"

go vet ./...
go test ./... -race -p 2 -parallel 1 -failfast --coverprofile=cover.out.tmp

go tool cover -html=cover.out.tmp -o cover.html
cat cover.out.tmp | grep -v \
                         -e ".pb.go" \
                         -e ".pb.gw.go" \
                         -e "/nulls/" > cover.out
go tool cover -func cover.out