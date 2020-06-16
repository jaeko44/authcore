#!/bin/sh

set -e
set -o xtrace

BASEDIR=$( cd "$(dirname "$0")"/.. ; pwd -P )
if [ -z "$BASEDIR" ]
then
    echo "cannot determine base path"
    exit 255
fi

VERSION=$(git describe --always --match 'v*')

if [ $VERSION != v* ]
then
    VERSION=0.1.0+git.$VERSION
fi

export GOBIN="$BASEDIR"/build

rm -fr "$GOBIN"/*
cd "$BASEDIR"
go install -ldflags "-X authcore.io/authcore/internal/server.buildVersion=${VERSION}" authcore.io/authcore/cmd/...