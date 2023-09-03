package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	pb "visualjerk.de/challengers/grpc"
)

type Subscriber struct {
	stream *pb.Game_GameEventsServer
	done   chan bool
}

func newSubscriber(stream *pb.Game_GameEventsServer) *Subscriber {
	return &Subscriber{stream, make(chan bool)}
}

func (s *Subscriber) send(event *pb.GameEvent) error {
	if err := (*s.stream).Send(event); err != nil {
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

func (g *GameEvents) addSubscriber(stream *pb.Game_GameEventsServer) error {
	subscriber := newSubscriber(stream)
	id := uuid.NewString()
	g.subscribers[id] = subscriber
	fmt.Printf("added subscriber with id %s\n", id)

	defer delete(g.subscribers, id)
	defer fmt.Printf("removed subscriber with id %s\n", id)

	// Send events that have been published so far
	subscriber.sendMany(g.events)

	// Keep stream open until subscriber is done
	<-subscriber.done
	return nil
}

func (g *GameEvents) publish(event *pb.GameEvent) {
	g.events = append(g.events, event)

	for id, subscriber := range g.subscribers {
		fmt.Printf("notify subscriber with id %s\n", id)
		go subscriber.send(event)
	}
}

type Player struct {
	id   string
	name string
}

type Game struct {
	id      string
	seats   int
	players map[string]*Player
	events  *GameEvents
}

func NewGame(id string, seats int) *Game {
	return &Game{
		id:      id,
		seats:   seats,
		players: map[string]*Player{},
		events:  newGameEvents(),
	}
}

func (g *Game) HandlePlayerAction(request *pb.PlayerActionRequest) (*pb.PlayerActionResponse, error) {
	event, error := g.getPlayerActionEvent(request)
	if error != nil {
		return nil, error
	}

	g.addEvent(event)

	response := &pb.PlayerActionResponse{
		Response: &pb.PlayerActionResponse_Success{},
	}
	return response, nil
}

func (g *Game) Subscribe(stream *pb.Game_GameEventsServer) error {
	return g.events.addSubscriber(stream)
}

func (g *Game) addEvent(event *pb.GameEvent) {
	g.events.publish(event)
}

func (g *Game) getPlayerActionEvent(request *pb.PlayerActionRequest) (*pb.GameEvent, error) {
	event := &pb.GameEvent{
		Id:      uuid.NewString(),
		Date:    time.Now().Format(time.RFC3339Nano),
		Message: nil,
	}
	switch message := request.Message.(type) {
	case *pb.PlayerActionRequest_PlayerJoin:
		event.Message = &pb.GameEvent_PlayerJoined{
			PlayerJoined: &pb.PlayerJoined{
				Id:   uuid.NewString(),
				Name: message.PlayerJoin.Name,
			},
		}
	case *pb.PlayerActionRequest_PlayerLeave:
		event.Message = &pb.GameEvent_PlayerLeft{
			PlayerLeft: &pb.PlayerLeft{
				Id:   message.PlayerLeave.PlayerId,
				Name: message.PlayerLeave.PlayerId,
			},
		}
	default:
		return nil, fmt.Errorf("unknown player action")
	}
	return event, nil
}

type GameServer struct {
	pb.GameServer
	games map[string]*Game
}

func NewServer() *GameServer {
	s := &GameServer{
		games: map[string]*Game{},
	}
	return s
}

func (s *GameServer) PlayerAction(
	context context.Context,
	request *pb.PlayerActionRequest,
) (*pb.PlayerActionResponse, error) {
	game := s.games[request.GameId]

	if game == nil {
		return nil, fmt.Errorf("game with id %s not found", request.GameId)
	}

	return game.HandlePlayerAction(request)
}

func (s *GameServer) GameEvents(
	request *pb.GameEventsSubscriptionRequest,
	stream pb.Game_GameEventsServer,
) error {
	game := s.games[request.GameId]

	if game == nil {
		return fmt.Errorf("game with id %s not found", request.GameId)
	}

	error := game.Subscribe(&stream)
	return error
}

func (s *GameServer) CreateGame(
	context context.Context,
	request *pb.CreateGameRequest,
) (*pb.CreateGameResponse, error) {
	id := uuid.NewString()
	game := NewGame(id, 4)
	s.games[id] = game
	fmt.Printf("created game with id %s\n", id)
	return &pb.CreateGameResponse{Id: id}, nil
}
