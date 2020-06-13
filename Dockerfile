FROM golang:1.13

RUN apt-get update && \
    apt-get install -y \
    build-essential \
    git \
    unzip \
    mariadb-client && \
    apt-get clean

# Install protoc
RUN curl -s -L https://github.com/protocolbuffers/protobuf/releases/download/v3.6.1/protoc-3.6.1-linux-x86_64.zip -o /tmp/protoc.zip && \
    unzip /tmp/protoc.zip -d /usr/local -x readme.txt && \
    rm -f /tmp/protoc.zip

ENV GO111MODULE=on

# Install developer tools
RUN go get -v \
    github.com/cespare/reflex \
    github.com/golang/protobuf/protoc-gen-go@v1.3.1 \
    github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.8.5 \
    github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.8.5 \
    github.com/amacneil/dbmate@v1.4.1

# Install node.js
RUN curl -s -L https://deb.nodesource.com/setup_12.x | bash
RUN apt-get install nodejs
RUN npm install yarn -g
