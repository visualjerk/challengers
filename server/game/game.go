package game

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	pb "visualjerk.de/challengers/grpc"
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

type GameServer struct {
	pb.GameServer
	events *GameEvents
}

func NewServer() *GameServer {
	s := &GameServer{
		events: newGameEvents(),
	}
	return s
}

func (s *GameServer) addEvent(event *pb.GameEvent) {
	s.events.publish(event)
}

func (s *GameServer) PlayerAction(
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

func (s *GameServer) GameEvents(
	request *pb.GameEventsSubscriptionRequest,
	stream pb.Game_GameEventsServer,
) error {
	return s.events.addSubscriber(stream)
}
