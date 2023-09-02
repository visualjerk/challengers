package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "visualjerk.de/challengers/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type gameServer struct {
	pb.GameServer
}

func newServer() *gameServer {
	s := &gameServer{}
	return s
}

func main() {
	flag.Parse()
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGameServer(grpcServer, newServer())

	fmt.Printf("Starting game server at: http://localhost:%d", *port)
	grpcServer.Serve(listener)
}
