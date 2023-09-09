# Challengers Server

## Development Setup

- [Prerequisites for gRPC / Protobuf generation](https://grpc.io/docs/languages/go/quickstart/#prerequisites)

## Generate gRPC Code

First ensure that you set a environment variable to the root path of this repository.

Example:

```bash
export CHALLENGERS_ROOT="/Users/myname/Development/challengers"
```

Generate code with:

```bash
go generate
```
