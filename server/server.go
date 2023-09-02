package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"

	pb "visualjerk.de/challengers/grpc"
)

var (
	host = flag.String("host", "0.0.0.0", "The server host")
	port = flag.Int("port", 50051, "The server port")
)

type Subscriber struct {
	stream pb.Game_GameEventsServer
	done   chan bool
}

func newSubscriber(stream pb.Game_GameEventsServer) *Subscriber {
	return &Subscriber{stream, make(chan bool)}
}

func (s *Subscriber) send(event *pb.GameEvent) error {
	if err := s.stream.Send(event); err != nil {
		s.done <- true
		return err
	}
	return nil
}

func (s *Subscriber) sendMany(events []*pb.GameEvent) error {
	for _, event := range events {
		if err := s.send(event); err != nil {
			return err
		}
	}
	return nil
}

type GameEvents struct {
	subscribers map[string]*Subscriber
	events      []*pb.GameEvent
}

func newGameEvents() *GameEvents {
	return &GameEvents{
		map[string]*Subscriber{},
		[]*pb.GameEvent{},
	}
}

func (s *GameEvents) addSubscriber(stream pb.Game_GameEventsServer) error {
	subscriber := newSubscriber(stream)
	id := uuid.NewString()
	s.subscribers[id] = subscriber
	fmt.Printf("added subscriber with id %s\n", id)

	defer delete(s.subscribers, id)
	defer fmt.Printf("removed subscriber with id %s\n", id)

	// Send events that have been published so far
	subscriber.sendMany(s.events)

	// Keep stream open until subscriber is done
	<-subscriber.done
	return nil
}

func (s *GameEvents) publish(event *pb.GameEvent) {
	s.events = append(s.events, event)

	for id, subscriber := range s.subscribers {
		fmt.Printf("notify subscriber with id %s\n", id)
		go subscriber.send(event)
	}
}

type gameServer struct {
	pb.GameServer
	events *GameEvents
}

func newServer() *gameServer {
	s := &gameServer{
		events: newGameEvents(),
	}
	return s
}

func (s *gameServer) addEvent(event *pb.GameEvent) {
	s.events.publish(event)
}

func (s *gameServer) PlayerAction(
	context context.Context,
	request *pb.PlayerActionRequest,
) (*pb.PlayerActionResponse, error) {
	s.addEvent(&pb.GameEvent{
		Id:   uuid.NewString(),
		Date: "DATE",
		Message: &pb.GameEvent_PlayerJoined{
			PlayerJoined: &pb.PlayerJoined{
				Id:   uuid.NewString(),
				Name: request.GetPlayerJoin().GetName(),
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
	return s.events.addSubscriber(stream)
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
