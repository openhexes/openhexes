package config

import (
	"fmt"
	"time"
)

type Auth struct {
	Google             GoogleAuth  `envPrefix:"GOOGLE__"`
	Storage            AuthStorage `envPrefix:"STORAGE__"`
	Owners             Owners      `envPrefix:"OWNERS__"`
	PreActivatedEmails []string    `env:"PRE_ACTIVATED_EMAILS"`
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

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func fillPicture(c *GoogleClaims) *GoogleClaims {
	if c.Picture == "" {
		c.Picture = fmt.Sprintf("https://i.pravatar.cc/300?u=%s", c.Email)
	}
	return c
}
