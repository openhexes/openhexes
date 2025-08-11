package config

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Test struct {
	Enabled  bool `env:"ENABLED"`
	ID       string
	Tokens   Tokens
	Accounts map[string]*GoogleClaims
}

type Tokens struct {
	Owner      string
	Unverified string
	Alfa       string
	Bravo      string
}

func (cfg *Config) setUpTests(ctx context.Context) error {
	if !cfg.Test.Enabled {
		return nil
	}

	cfg.Test.ID = fmt.Sprintf(
		"%s-%s",
		strings.ReplaceAll(time.Now().Format(time.TimeOnly), ":", "-"),
		uuid.NewString()[24:],
	)
	cfg.Auth.Owners.Emails = []string{"owner@test.com"}

	cfg.Test.Tokens = Tokens{
		Owner:      "owner",
		Unverified: "unverified",
		Alfa:       "alfa",
		Bravo:      "bravo",
	}

	cfg.Test.Accounts = map[string]*GoogleClaims{
		cfg.Test.Tokens.Owner: fillPicture(&GoogleClaims{
			Email:         "owner@test.com",
			EmailVerified: true,
			Name:          "Test Owner",
		}),
		cfg.Test.Tokens.Unverified: fillPicture(&GoogleClaims{
			Email: "unverified@test.com",
			Name:  "Test Unverified",
		}),
		cfg.Test.Tokens.Alfa: fillPicture(&GoogleClaims{
			Email:         "alfa@test.com",
			EmailVerified: true,
			Name:          "Test Alfa",
		}),
		cfg.Test.Tokens.Bravo: fillPicture(&GoogleClaims{
			Email:         "bravo@test.com",
			EmailVerified: true,
			Name:          "Test Bravo",
		}),
	}

	for _, a := range cfg.Test.Accounts {
		cfg.Auth.PreActivatedEmails = append(cfg.Auth.PreActivatedEmails, a.Email)
	}

	return nil
}
