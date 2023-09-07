package account

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"

	pb "visualjerk.de/challengers/grpc"
)

type Account struct {
	Token string
	Id    string
	Name  string
}

func NewAccount(token string, id string, name string) *Account {
	return &Account{
		token,
		id,
		name,
	}
}

type AccountServer struct {
	pb.AccountServer
	accounts        map[string]*Account
	accountsByToken map[string]*Account
}

func NewAccountServer() *AccountServer {
	s := &AccountServer{
		accounts:        map[string]*Account{},
		accountsByToken: map[string]*Account{},
	}
	return s
}

func (s *AccountServer) CreateAccount(
	context context.Context,
	request *pb.CreateAccountRequest,
) (*pb.CreateAccountResponse, error) {
	token := uuid.NewString()
	id := uuid.NewString()
	account := NewAccount(token, id, request.Name)

	s.accounts[id] = account

	// TODO: Encrypt token
	s.accountsByToken[token] = account
	fmt.Printf("created account with id %s\n", id)

	return &pb.CreateAccountResponse{Token: token}, nil
}

func (s *AccountServer) GetAccount(context context.Context) (*Account, error) {
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
