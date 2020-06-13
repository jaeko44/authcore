#!/bin/sh

set -e
set -o xtrace

BASEDIR=$( cd "$(dirname "$0")"/.. ; pwd -P )
API_DIR="$BASEDIR"/api
DOCS_DIR="$BASEDIR"/docs
BUILD_DIR="$BASEDIR"/build/docs

mkdir -p "$BUILD_DIR"
cd "$DOCS_DIR"
yarn install
yarn redoc-cli bundle "$API_DIR"/authapi/authcore.swagger.json -o "$BUILD_DIR"/authapi.html -t docs.hbs --title "AuthAPI documentation" --options.hideHostname true
yarn redoc-cli bundle "$API_DIR"/managementapi/management.swagger.json -o "$BUILD_DIR"/management.html -t docs.hbs --title "ManagementAPI documentation" --options.hideHostname true
cp "$DOCS_DIR"/index.html "$BUILD_DIR"
