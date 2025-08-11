package test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/server"
	"github.com/openhexes/proto/game/v1/gamev1connect"
	"github.com/openhexes/proto/iam/v1/iamv1connect"
	"github.com/stretchr/testify/require"
)

type Env struct {
	*testing.T
	Require *Assertions

	Config  *config.Config
	Auth    *auth.Controller
	Server  *server.Server
	Clients *Clients
}

func NewEnv(t *testing.T, configOptions ...config.Option) *Env {
	t.Parallel()

	cfg, err := config.New(
		t.Context(),
		append(
			configOptions,
			config.WithTestMode(),
			config.WithRandomServerAddress(),
		)...,
	)
	require.NoError(t, err, "creating test config")

	auth := auth.NewController(cfg)

	srv, err := server.New(cfg, auth)
	require.NoError(t, err, "creating server")

	e := &Env{
		T:       t,
		Require: &Assertions{require.New(t)},
		Config:  cfg,
		Auth:    auth,
		Server:  srv,
		Clients: &Clients{},
	}

	e.Require.NoError(e.Config.SetUp(e.Context()))
	e.Require.NoError(e.Server.Init(), "initializing server")

	// Use a separate context for the server to prevent early shutdown
	serverCtx, serverCancel := context.WithCancel(context.Background())
	t.Cleanup(serverCancel)

	go func() {
		if err := e.Server.Run(serverCtx); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	addr := fmt.Sprintf("http://%s", cfg.Server.Address)
	opts := []connect.ClientOption{
		connect.WithGRPC(),
	}
	e.Clients.IAM = iamv1connect.NewIAMServiceClient(http.DefaultClient, addr, opts...)
	e.Clients.Game = gamev1connect.NewGameServiceClient(http.DefaultClient, addr, opts...)

	// todo: replace with /ping check
	// Give the server a moment to start up
	time.Sleep(10 * time.Millisecond)

	t.Cleanup(e.TearDown)

	return e
}

func (e *Env) TearDown() {
	e.Require.NoError(e.Config.TearDown(context.Background()))
}

func (e *Env) Run(name string, fn func(e *Env)) {
	e.Helper()
	e.T.Run(name, func(t *testing.T) {
		e.Helper()

		prevT := e.T
		prevRequire := e.Require

		e.T = t
		e.Require = &Assertions{require.New(t)}

		defer func() {
			e.T = prevT
			e.Require = prevRequire
		}()

		fn(e)
	})
}
