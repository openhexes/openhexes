package main

import (
	"context"
	"log"
	"os"

	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/server"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	var srvErr error

	cfg, err := config.New(ctx)
	if err != nil {
		log.Fatalf("loading config: %s", err)
	}
	if err = cfg.SetUp(ctx); err != nil {
		zap.L().Fatal("failed to set up: %s", zap.Error(err))
	}
	defer func() {
		if err = cfg.TearDown(ctx); err != nil {
			zap.L().Error("failed to tear down: %s", zap.Error(err))
		}
		if srvErr != nil {
			os.Exit(1)
		}
	}()

	var srv *server.Server
	srv, srvErr = server.New(cfg, auth.NewController(cfg))
	if srvErr != nil {
		zap.L().Error("creating server: %s", zap.Error(err))
	} else {
		if srvErr = srv.Init(); srvErr != nil {
			zap.L().Error("initializing server", zap.Error(srvErr))
			return
		}
		if srvErr = srv.Run(ctx); srvErr != nil {
			zap.L().Error("running server: %s", zap.Error(srvErr))
		}
	}
}
