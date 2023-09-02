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

type subscriber struct {
	stream pb.Game_GameEventsServer
	done   chan bool
}

func newSubscriber(stream pb.Game_GameEventsServer) *subscriber {
	return &subscriber{stream, make(chan bool)}
}

func (s *subscriber) notify(event *pb.GameEvent) error {
	if err := s.stream.Send(event); err != nil {
		s.done <- true
		return err
	}
	return nil
}

type gameServer struct {
	pb.GameServer
	subscribers []*subscriber
}

func newServer() *gameServer {
	s := &gameServer{
		subscribers: []*subscriber{},
	}
	return s
}

func (s *gameServer) addEvent(event *pb.GameEvent) {
	for _, subscriber := range s.subscribers {
		go subscriber.notify(event)
	}
}

func (s *gameServer) PlayerAction(
	context context.Context,
	request *pb.PlayerActionRequest,
) (*pb.PlayerActionResponse, error) {
	s.addEvent(&pb.GameEvent{
		Id:   "ID",
		Date: "DATE",
		Message: &pb.GameEvent_PlayerJoined{
			PlayerJoined: &pb.PlayerJoined{
				Id:   request.GetPlayerJoined().Id,
				Name: request.GetPlayerJoined().Name,
			},
		},
	})

	response := &pb.PlayerActionResponse{
		Response: &pb.PlayerActionResponse_Success{},
	}
	return response, nil
}

func (s *gameServer) GameEvents(
	request *pb.GameEventsSubscriptionRequest,
	stream pb.Game_GameEventsServer,
) error {
	subscriber := newSubscriber(stream)
	// Subscribe this client
	s.subscribers = append(s.subscribers, subscriber)
	<-subscriber.done
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
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
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
		}),
	}

	fmt.Printf("Starting game server at: http://%s:%d\n", *host, *port)
	err := httpServer.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
