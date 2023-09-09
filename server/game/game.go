package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"

	"visualjerk.de/challengers/account"
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

func (g *Game) HandlePlayerAction(request *pb.PlayerActionRequest, player *Player) (*pb.PlayerActionResponse, error) {
	event, error := g.getPlayerActionEvent(request, player)
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

func (g *Game) getGameState() *pb.GameState {
	players := []*pb.Player{}

	for _, player := range g.players {
		players = append(players, &pb.Player{
			Id:   player.id,
			Name: player.name,
		})
	}

	return &pb.GameState{
		Players: players,
	}
}

func (g *Game) getPlayerActionEvent(request *pb.PlayerActionRequest, player *Player) (*pb.GameEvent, error) {
	event := &pb.GameEvent{
		Id:   uuid.NewString(),
		Date: time.Now().Format(time.RFC3339Nano),
	}
	switch request.Message.(type) {
	case *pb.PlayerActionRequest_PlayerJoin:
		g.players[player.id] = player

		event.Message = &pb.GameEvent_PlayerJoined{
			PlayerJoined: &pb.PlayerJoined{
				Player: &pb.Player{
					Id:   player.id,
					Name: player.name,
				},
			},
		}
	case *pb.PlayerActionRequest_PlayerLeave:
		leavingPlayer := g.players[player.id]
		if leavingPlayer == nil {
			return nil, fmt.Errorf("player is not in this game")
		}

		delete(g.players, leavingPlayer.id)

		event.Message = &pb.GameEvent_PlayerLeft{
			PlayerLeft: &pb.PlayerLeft{
				Player: &pb.Player{
					Id:   player.id,
					Name: player.name,
				},
			},
		}
	default:
		return nil, fmt.Errorf("unknown player action")
	}
	event.State = g.getGameState()
	return event, nil
}

type GameServer struct {
	pb.GameServer
	games         map[string]*Game
	accountServer *account.AccountServer
}

func NewServer(accountServer *account.AccountServer) *GameServer {
	s := &GameServer{
		games:         map[string]*Game{},
		accountServer: accountServer,
	}
	return s
}

func (s *GameServer) AddToGrpcServer(server *grpc.Server) {
	pb.RegisterGameServer(server, s)
}

func (s *GameServer) PlayerAction(
	context context.Context,
	request *pb.PlayerActionRequest,
) (*pb.PlayerActionResponse, error) {
	account, error := s.accountServer.GetAccount(context)
	if error != nil {
		return nil, error
	}

	game := s.games[request.GameId]

	if game == nil {
		return nil, fmt.Errorf("game with id %s not found", request.GameId)
	}

	return game.HandlePlayerAction(request, &Player{
		name: account.Name,
		id:   account.Id,
	})
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
	if _, error := s.accountServer.GetAccount(context); error != nil {
		return nil, error
	}

	id := uuid.NewString()
	game := NewGame(id, 4)
	s.games[id] = game
	fmt.Printf("created game with id %s\n", id)
	return &pb.CreateGameResponse{Id: id}, nil
}
