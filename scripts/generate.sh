#!/bin/sh

set -e
set -o xtrace

BASEDIR=$(dirname "$0")/..
PROTOC_INCLUDES="-I$BASEDIR/api -I/usr/local/include -I$BASEDIR/third_party/googleapis -I$BASEDIR/third_party/grpc-gateway"
GENERATED_DIR="$BASEDIR"/pkg/api
API_DIR="$BASEDIR"/api

protoc $PROTOC_INCLUDES --go_out=plugins=grpc,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/authapi/authcore.proto
protoc $PROTOC_INCLUDES --go_out=plugins=grpc,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/authapi/authcore_entity.proto
protoc $PROTOC_INCLUDES --grpc-gateway_out=logtostderr=true,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/authapi/authcore.proto
protoc $PROTOC_INCLUDES --swagger_out=logtostderr=true:"$API_DIR" "$API_DIR"/authapi/authcore.proto

protoc $PROTOC_INCLUDES --go_out=plugins=grpc,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/managementapi/management.proto
protoc $PROTOC_INCLUDES --go_out=plugins=grpc,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/managementapi/management_entity.proto
protoc $PROTOC_INCLUDES --grpc-gateway_out=logtostderr=true,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/managementapi/management.proto
protoc $PROTOC_INCLUDES --swagger_out=logtostderr=true:"$API_DIR" "$API_DIR"/managementapi/management.proto

protoc $PROTOC_INCLUDES --go_out=plugins=grpc,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/secretdgateway/secretdgateway.proto
protoc $PROTOC_INCLUDES --grpc-gateway_out=logtostderr=true,paths=source_relative:"$GENERATED_DIR" "$API_DIR"/secretdgateway/secretdgateway.proto
protoc $PROTOC_INCLUDES --swagger_out=logtostderr=true:"$API_DIR" "$API_DIR"/secretdgateway/secretdgateway.proto

go generate authcore.io/authcore/cmd/authcored/app