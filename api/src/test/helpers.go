package test

import (
	"connectrpc.com/connect"
	"github.com/openhexes/openhexes/api/src/config"
	gamev1 "github.com/openhexes/proto/game/v1"
	iamv1 "github.com/openhexes/proto/iam/v1"
	mapv1 "github.com/openhexes/proto/map/v1"
	worldv1 "github.com/openhexes/proto/world/v1"
)

type HelperCallConfig struct {
	RequestOptions    []config.RequestOption
	ExpectedErrorCode *connect.Code
}

type HelperCallOption func(*HelperCallConfig)

func WithToken(t string) HelperCallOption {
	return func(c *HelperCallConfig) {
		c.RequestOptions = append(c.RequestOptions, config.WithToken(t))
	}
}

func WithExpectedCode(code connect.Code) HelperCallOption {
	return func(c *HelperCallConfig) {
		c.ExpectedErrorCode = &code
	}
}

func getHelperCallConfig(opts ...HelperCallOption) *HelperCallConfig {
	var cfg HelperCallConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	return &cfg
}

func (e *Env) ResolveAccount(opts ...HelperCallOption) *iamv1.Account {
	e.Helper()
	cfg := getHelperCallConfig(opts...)

	request := connect.NewRequest(&iamv1.ResolveAccountRequest{})
	for _, opt := range cfg.RequestOptions {
		opt(request.Header())
	}

	response, err := e.Clients.IAM.ResolveAccount(e.Context(), request)
	e.Require.ErrorCode(err, cfg.ExpectedErrorCode)
	e.Require.NotNil(response.Msg.Account)
	return response.Msg.Account
}

func (e *Env) ListAccounts(opts ...HelperCallOption) []*iamv1.Account {
	e.Helper()
	cfg := getHelperCallConfig(opts...)

	request := connect.NewRequest(&iamv1.ListAccountsRequest{})
	for _, opt := range cfg.RequestOptions {
		opt(request.Header())
	}

	stream, err := e.Clients.IAM.ListAccounts(e.Context(), request)
	e.Require.ErrorCode(err, cfg.ExpectedErrorCode)

	accounts := make([]*iamv1.Account, 0, 100)
	for stream.Receive() {
		accounts = append(accounts, stream.Msg().GetAccounts()...)
	}
	e.Require.NoError(stream.Err(), "iterating stream")
	e.Require.NoError(stream.Close(), "closing stream")
	return accounts
}

func (e *Env) GetAccount(id string, opts ...HelperCallOption) *iamv1.Account {
	e.Helper()
	cfg := getHelperCallConfig(opts...)

	request := connect.NewRequest(&iamv1.GetAccountRequest{
		Id:        id,
		WithRoles: true,
	})
	for _, opt := range cfg.RequestOptions {
		opt(request.Header())
	}
	response, err := e.Clients.IAM.GetAccount(e.Context(), request)
	e.Require.ErrorCode(err, cfg.ExpectedErrorCode)
	e.Require.NotNil(response.Msg.Account)
	return response.Msg.Account
}

func (e *Env) UpdateAccountActivation(idToActivation map[string]bool, opts ...HelperCallOption) {
	e.Helper()
	cfg := getHelperCallConfig(opts...)

	request := connect.NewRequest(&iamv1.UpdateAccountActivationRequest{
		IdToActivation: idToActivation,
	})
	for _, opt := range cfg.RequestOptions {
		opt(request.Header())
	}

	_, err := e.Clients.IAM.UpdateAccountActivation(e.Context(), request)
	e.Require.ErrorCode(err, cfg.ExpectedErrorCode)
}

func (e *Env) GetSampleWorld(options *gamev1.GetSampleWorldRequest, opts ...HelperCallOption) *worldv1.World {
	e.Helper()
	cfg := getHelperCallConfig(opts...)

	request := connect.NewRequest(options)
	for _, opt := range cfg.RequestOptions {
		opt(request.Header())
	}

	stream, err := e.Clients.Game.GetSampleWorld(e.Context(), request)
	e.Require.ErrorCode(err, cfg.ExpectedErrorCode)

	var world *worldv1.World
	for stream.Receive() {
		response := stream.Msg()
		if response.World == nil {
			continue
		}
		if world == nil {
			world = response.World
			continue
		}
		if response.World.RenderingSpec != nil {
			world.RenderingSpec = response.World.RenderingSpec
		}
		if response.World.TerrainRegistry != nil {
			if world.TerrainRegistry == nil {
				world.TerrainRegistry = make(map[string]*mapv1.Terrain)
			}
			for k, v := range response.World.TerrainRegistry {
				world.TerrainRegistry[k] = v
			}
		}
		if len(response.World.Layers) > 0 {
			if len(world.Layers) == 0 {
				world.Layers = make([]*mapv1.Grid, len(response.World.Layers))
			}
			for i, layer := range response.World.Layers {
				if layer != nil {
					if world.Layers[i] == nil {
						world.Layers[i] = layer
					} else {
						// Merge segment rows
						world.Layers[i].SegmentRows = append(world.Layers[i].SegmentRows, layer.SegmentRows...)
					}
				}
			}
		}
	}
	e.Require.NoError(stream.Err())
	e.Require.NoError(stream.Close())
	e.Require.NotNil(world)
	return world
}
