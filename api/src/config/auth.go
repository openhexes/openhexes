package config

import "time"

type Auth struct {
	Google  GoogleAuth  `envPrefix:"GOOGLE__"`
	Storage AuthStorage `envPrefix:"STORAGE__"`
	Owners  Owners      `envPrefix:"OWNERS__"`
}

type GoogleAuth struct {
	ClientID string `env:"CLIENT_ID"`
}

type AuthStorage struct {
	MaxSize int           `envDefault:"256"`
	TTL     time.Duration `envDefault:"1h"`
}

type Owners struct {
	Emails []string `env:"EMAILS"`
	Roles  []string `env:"ROLES" envDefault:"owner"`
}
