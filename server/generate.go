package main

//go:generate protoc --go_out=./grpc --go_opt=paths=source_relative --go-grpc_out=./grpc --go-grpc_opt=paths=source_relative --proto_path=$CHALLENGERS_ROOT/proto account.proto game.proto
