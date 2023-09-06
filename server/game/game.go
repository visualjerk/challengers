package game

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

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

func (g *Game) HandlePlayerAction(request *pb.PlayerActionRequest, playerId string) (*pb.PlayerActionResponse, error) {
	event, error := g.getPlayerActionEvent(request, playerId)
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

func (g *Game) getPlayerActionEvent(request *pb.PlayerActionRequest, playerId string) (*pb.GameEvent, error) {
	event := &pb.GameEvent{
		Id:      uuid.NewString(),
		Date:    time.Now().Format(time.RFC3339Nano),
		Message: nil,
	}
	switch message := request.Message.(type) {
	case *pb.PlayerActionRequest_PlayerJoin:
		player := &Player{
			id:   playerId,
			name: message.PlayerJoin.Name,
		}
		g.players[playerId] = player

		event.Message = &pb.GameEvent_PlayerJoined{
			PlayerJoined: &pb.PlayerJoined{
				Id:   player.id,
				Name: player.name,
			},
		}
	case *pb.PlayerActionRequest_PlayerLeave:
		if playerId != message.PlayerLeave.PlayerId {
			return nil, fmt.Errorf("unauthorized action")
		}
		player := g.players[playerId]
		if player == nil {
			return nil, fmt.Errorf("player is not in this game")
		}

		g.players[playerId] = nil

		event.Message = &pb.GameEvent_PlayerLeft{
			PlayerLeft: &pb.PlayerLeft{
				Id:   player.id,
				Name: player.name,
			},
		}
	default:
		return nil, fmt.Errorf("unknown player action")
	}
	return event, nil
}

type GameServer struct {
	pb.GameServer
	games           map[string]*Game
	accounts        map[string]*Account
	accountsByToken map[string]*Account
}

func NewServer() *GameServer {
	s := &GameServer{
		games:           map[string]*Game{},
		accounts:        map[string]*Account{},
		accountsByToken: map[string]*Account{},
	}
	return s
}

func (s *GameServer) PlayerAction(
	context context.Context,
	request *pb.PlayerActionRequest,
) (*pb.PlayerActionResponse, error) {
	account, error := s.getAccount(context)
	if error != nil {
		return nil, error
	}

	game := s.games[request.GameId]

	if game == nil {
		return nil, fmt.Errorf("game with id %s not found", request.GameId)
	}

	return game.HandlePlayerAction(request, account.id)
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
	if _, error := s.getAccount(context); error != nil {
		return nil, error
	}

	id := uuid.NewString()
	game := NewGame(id, 4)
	s.games[id] = game
	fmt.Printf("created game with id %s\n", id)
	return &pb.CreateGameResponse{Id: id}, nil
}

func (s *GameServer) getAccount(context context.Context) (*Account, error) {
	authdata := metadata.ValueFromIncomingContext(context, "authorization")

	if len(authdata) < 1 {
		return nil, fmt.Errorf("missing auth token")
	}

	account := s.accountsByToken[authdata[0]]

	if account == nil {
		return nil, fmt.Errorf("account not found")
	}

	return account, nil
}

type Account struct {
	token string
	id    string
}

func NewAccount(token string, id string) *Account {
	return &Account{
		token,
		id,
	}
}

func (s *GameServer) CreateAccount(
	context context.Context,
	request *pb.CreateAccountRequest,
) (*pb.CreateAccountResponse, error) {
	token := uuid.NewString()
	id := uuid.NewString()
	account := NewAccount(token, id)

	s.accounts[id] = account

	// TODO: Encrypt token
	s.accountsByToken[token] = account
	fmt.Printf("created account with id %s\n", id)

	return &pb.CreateAccountResponse{Token: token}, nil
}
