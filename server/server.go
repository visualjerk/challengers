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

type subscribers struct {
	all    map[int]*subscriber
	lastId int
}

func newSubscribers() *subscribers {
	return &subscribers{
		map[int]*subscriber{},
		0,
	}
}

func (s *subscribers) add(stream pb.Game_GameEventsServer) error {
	subscriber := newSubscriber(stream)
	id := s.lastId + 1
	s.all[id] = subscriber
	fmt.Printf("added subscriber with id %d\n", id)

	s.lastId = id

	// Remove once it is done
	<-subscriber.done
	delete(s.all, id)
	fmt.Printf("removed subscriber with id %d\n", id)

	return nil
}

func (s *subscribers) notify(event *pb.GameEvent) {
	for id, subscriber := range s.all {
		fmt.Printf("notify subscriber with id %d\n", id)
		go subscriber.notify(event)
	}
}

type gameServer struct {
	pb.GameServer
	subscribers *subscribers
}

func newServer() *gameServer {
	s := &gameServer{
		subscribers: newSubscribers(),
	}
	return s
}

func (s *gameServer) addEvent(event *pb.GameEvent) {
	s.subscribers.notify(event)
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
	s.subscribers.add(stream)
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
