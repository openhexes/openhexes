package config

import (
	"net/http"

	connectcors "connectrpc.com/cors"
	"github.com/rs/cors"
)

func (cfg *Config) AddCORS(h http.Handler) http.Handler {
	middleware := cors.New(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   connectcors.AllowedMethods(),
		AllowedHeaders:   connectcors.AllowedHeaders(),
		ExposedHeaders:   connectcors.ExposedHeaders(),
		AllowCredentials: true,
	})
	return middleware.Handler(h)
}
