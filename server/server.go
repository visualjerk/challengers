package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	pb "visualjerk.de/challengers/grpc"
)

var (
	host = flag.String("host", "0.0.0.0", "The server host")
	port = flag.Int("port", 50051, "The server port")
)

type gameServer struct {
	pb.GameServer
}

func newServer() *gameServer {
	s := &gameServer{}
	return s
}

func (s *gameServer) PlayerAction(
	context context.Context,
	request *pb.PlayerActionRequest,
) (*pb.PlayerActionResponse, error) {
	response := &pb.PlayerActionResponse{
		Response: &pb.PlayerActionResponse_Success{},
	}
	return response, nil
}

func (s *gameServer) GameEvents(
	request *pb.GameEventsSubscriptionRequest,
	stream pb.Game_GameEventsServer,
) error {
	for i := 0; i < 10; i++ {
		event := &pb.GameEvent{
			Id:   fmt.Sprintf("%d", i),
			Date: "",
			Message: &pb.GameEvent_PlayerJoined{
				PlayerJoined: &pb.PlayerJoined{
					Id:   fmt.Sprintf("PlayerID%d", i),
					Name: "Han Solo",
				},
			},
		}

		if err := stream.Send(event); err != nil {
			return err
		}
	}
	return nil
}

func enableCors(resp *http.ResponseWriter) {
	(*resp).Header().Add("Access-Control-Allow-Origin", "*")
	(*resp).Header().Add("Access-Control-Allow-Methods", "DELETE, POST, GET, OPTIONS")
	(*resp).Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, x-grpc-web")
}

func main() {
	flag.Parse()

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	pb.RegisterGameServer(grpcServer, newServer())

	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	httpServer := &http.Server{
		Addr: fmt.Sprintf("%s:%d", *host, *port),
	}
	httpServer.Handler = http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
		enableCors(&resp)
		if req.Method == "OPTIONS" {
			resp.WriteHeader(http.StatusOK)
			return
		}

		if wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.ServeHTTP(resp, req)
			return
		}
		// Fall back to other servers.
		http.DefaultServeMux.ServeHTTP(resp, req)
	})

	fmt.Printf("Starting game server at: http://%s:%d", *host, *port)
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
