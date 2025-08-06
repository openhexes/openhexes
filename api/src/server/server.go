package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/services/iam"
	"github.com/openhexes/proto/iam/v1/iamv1connect"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type Server struct {
	*http.Server

	cfg      *config.Config
	listener net.Listener
}

func New(cfg *config.Config, auth *auth.Controller) (*Server, error) {
	mux := http.NewServeMux()

	otel, err := otelconnect.NewInterceptor()
	if err != nil {
		return nil, fmt.Errorf("initializing OpenTelemetry interceptor: %w", err)
	}

	interceptors := connect.WithInterceptors(
		otel,
		auth,
		NewLoggingInterceptor(cfg),
	)

	path, handler := iamv1connect.NewIAMServiceHandler(iam.New(cfg, auth), interceptors)
	mux.Handle(path, handler)

	mux.Handle("/ping", &Ponger{})

	ui, err := GetUIHandler()
	if err != nil {
		return nil, fmt.Errorf("configuring UI handler: %w", err)
	}
	mux.Handle("/", ui)

	return &Server{
		cfg: cfg,
		Server: &http.Server{
			Addr:    cfg.Server.Address,
			Handler: cfg.AddCORS(otelhttp.NewHandler(h2c.NewHandler(mux, &http2.Server{}), "/")),
		},
	}, nil
}

func (s *Server) Init() error {
	var err error
	s.listener, err = net.Listen("tcp", s.cfg.Server.Address)
	if err != nil {
		return fmt.Errorf("listening on %q: %w", s.cfg.Server.Address, err)
	}
	s.cfg.Server.Address = s.listener.Addr().String()
	return nil
}

func (s *Server) Run(ctx context.Context) (err error) {
	log := config.GetLogger(ctx)

	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := s.cfg.SetupTelemetry(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(ctx))
	}()

	// Actually start the server.
	log.Info("starting server", zap.String("address", s.cfg.Server.Address))
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- s.Server.Serve(s.listener)
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = s.Server.Shutdown(ctx)
	log.Info("server shutdown complete", zap.Error(err))
	return
}
