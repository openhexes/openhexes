package config

import "time"

type Auth struct {
	Google  GoogleAuth  `envPrefix:"GOOGLE__"`
	Storage AuthStorage `envPrefix:"STORAGE__"`
}

type GoogleAuth struct {
	ClientID string `env:"CLIENT_ID"`
}

type AuthStorage struct {
	MaxSize int           `envDefault:"256"`
	TTL     time.Duration `envDefault:"12h"`
}
