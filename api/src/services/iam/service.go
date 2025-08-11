package iam

import (
	"context"
	"fmt"
	"slices"

	"connectrpc.com/connect"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/conv"
	"github.com/openhexes/openhexes/api/src/db"
	v1 "github.com/openhexes/proto/iam/v1"
	v1connect "github.com/openhexes/proto/iam/v1/iamv1connect"
)

type Service struct {
	v1connect.UnimplementedIAMServiceHandler

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
	caller := auth.AccountFromContext(ctx)
	return connect.NewResponse(&v1.ResolveAccountResponse{Account: conv.AccountToProto(caller)}), nil
}

func (svc *Service) ListAccounts(ctx context.Context, request *connect.Request[v1.ListAccountsRequest], stream *connect.ServerStream[v1.ListAccountsResponse]) error {
	const chunkSize = 1000
	var (
		result []db.Account
		err    error
	)
	err = svc.cfg.Postgres.Tx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		result, err = q.ListAccounts(ctx, pgtype.Bool{Bool: request.Msg.GetActive(), Valid: request.Msg.Active != nil})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	for chunk := range slices.Chunk(result, chunkSize) {
		response := &v1.ListAccountsResponse{
			Accounts: make([]*v1.Account, 0, chunkSize),
		}
		for _, m := range chunk {
			response.Accounts = append(response.Accounts, conv.AccountToProto(&m))
		}
		if err = stream.Send(response); err != nil {
			return err
		}
	}
	return nil
}

func (svc *Service) GetAccount(ctx context.Context, request *connect.Request[v1.GetAccountRequest]) (*connect.Response[v1.GetAccountResponse], error) {
	var (
		result db.Account
		roles  []string
	)

	id, err := uuid.Parse(request.Msg.Id)
	if err != nil {
		return nil, InvalidID(request.Msg.Id, err)
	}

	err = svc.cfg.Postgres.Tx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		result, err = q.GetAccountByID(ctx, id)
		if err != nil {
			return fmt.Errorf("fetching account: %q: %w", id, err)
		}

		if request.Msg.WithRoles {
			roles, err = q.ListGrants(ctx, db.ListGrantsParams{AccountID: id})
			if err != nil {
				return fmt.Errorf("fetching account roles: %q: %w", id, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	account := conv.AccountToProto(&result)
	account.Roles = roles
	return connect.NewResponse(&v1.GetAccountResponse{Account: account}), nil
}

func (svc *Service) UpdateAccountActivation(ctx context.Context, request *connect.Request[v1.UpdateAccountActivationRequest]) (*connect.Response[v1.UpdateAccountActivationResponse], error) {
	if len(request.Msg.IdToActivation) == 0 {
		return nil, ErrIdToActivationRequired
	}

	toActivate := mapset.NewSetWithSize[uuid.UUID](len(request.Msg.IdToActivation))
	toDeactivate := mapset.NewSetWithSize[uuid.UUID](len(request.Msg.IdToActivation))
	for k, v := range request.Msg.IdToActivation {
		id, err := uuid.Parse(k)
		if err != nil {
			return nil, InvalidID(k, err)
		}
		if v {
			toActivate.Add(id)
		} else {
			toDeactivate.Add(id)
		}
	}

	caller := auth.AccountFromContext(ctx)

	err := svc.cfg.Postgres.Tx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		roles, err := q.ListGrants(ctx, db.ListGrantsParams{AccountID: caller.ID, RoleIds: []string{config.RoleOwner}})
		if err != nil {
			return fmt.Errorf("fetching account roles: %q: %w", caller.ID, err)
		}
		if len(roles) == 0 {
			return ErrDenied
		}

		if !toActivate.IsEmpty() {
			err = q.UpdateAccountActivation(ctx, db.UpdateAccountActivationParams{Ids: toActivate.ToSlice(), Active: true})
			if err != nil {
				return fmt.Errorf("activating %d account(s): %w", toActivate.Cardinality(), err)
			}
		}
		if !toDeactivate.IsEmpty() {
			err = q.UpdateAccountActivation(ctx, db.UpdateAccountActivationParams{Ids: toDeactivate.ToSlice(), Active: false})
			if err != nil {
				return fmt.Errorf("deactivating %d account(s): %w", toActivate.Cardinality(), err)
			}
		}
		return nil
	})
	return connect.NewResponse(&v1.UpdateAccountActivationResponse{}), err
}
