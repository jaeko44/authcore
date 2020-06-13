#!/bin/sh

set -e
set -o xtrace

BASEDIR=$( cd "$(dirname "$0")"/.. ; pwd -P )
if [ -z "$BASEDIR" ]
then
    echo "cannot determine base path"
    exit 255
fi

export GOBIN="$BASEDIR"/build

rm -fr "$GOBIN"/*
cd "$BASEDIR"
go install authcore.io/authcore/cmd/...