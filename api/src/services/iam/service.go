package iam

import (
	"context"

	"connectrpc.com/connect"
	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/conv"
	v1 "github.com/openhexes/proto/iam/v1"
	"github.com/openhexes/proto/iam/v1/iamv1connect"
)

type Service struct {
	iamv1connect.UnimplementedIAMServiceHandler

	cfg  *config.Config
	auth *auth.Controller
}

func New(cfg *config.Config, auth *auth.Controller) *Service {
	return &Service{
		cfg:  cfg,
		auth: auth,
	}
}

func (svc *Service) ResolveAccount(ctx context.Context, request *connect.Request[v1.ResolveAccountRequest]) (*connect.Response[v1.ResolveAccountResponse], error) {
	account, err := svc.auth.AccountFromRequestHeader(ctx, request.Header())
	return connect.NewResponse(&v1.ResolveAccountResponse{
		Account: conv.AccountToProto(account),
	}), err
}

func (svc *Service) ListAccounts(ctx context.Context, request *connect.Request[v1.ListAccountsRequest], stream *connect.ServerStream[v1.ListAccountsResponse]) error {
	return nil
}
