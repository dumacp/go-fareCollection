#!/bin/bash

#protoc -I=. -I=$GOPATH/src --gogoslick_out=. messages.proto
protoc -I=./ -I=$GOPATH/src --go_out=$GOPATH/src ./messages.proto
#protoc -I=./ -I=$GOPATH/src --go-grpc_out=$GOPATH/src ./messages.proto
protoc -I=. -I=$GOPATH/src --gograin_out=. messages.proto
