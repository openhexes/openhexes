package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"cloud.google.com/go/auth/credentials/idtoken"
	"connectrpc.com/connect"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/jackc/pgx/v5"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/db"
	"go.uber.org/zap"
)

type contextKey string

const (
	ContextKey contextKey = "account"
)

type Controller struct {
	cfg   *config.Config
	cache *expirable.LRU[string, *db.Account]
}

type GoogleClaims struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewController(cfg *config.Config) *Controller {
	return &Controller{
		cfg:   cfg,
		cache: expirable.NewLRU[string, *db.Account](cfg.Auth.Storage.MaxSize, nil, cfg.Auth.Storage.TTL),
	}
}

func AccountFromContext(ctx context.Context) *db.Account {
	return ctx.Value(ContextKey).(*db.Account)
}

func (c *Controller) AccountFromRequestHeader(ctx context.Context, header http.Header) (*db.Account, error) {
	log := config.GetLogger(ctx)
	log.Info("request:header", zap.String("header", fmt.Sprintf("%s", header)))

	cookie, err := (&http.Request{Header: header}).Cookie("hexes.auth.cookie")
	if err != nil {
		return nil, connect.NewError(
			connect.CodeInvalidArgument,
			fmt.Errorf("parsing authentication cookie: %w", err),
		)
	}

	bh := sha256.Sum256([]byte(cookie.Value))
	sh := hex.EncodeToString([]byte(bh[:]))

	account, ok := c.cache.Get(sh)
	if !ok {
		log.Debug("resolving new credentials", zap.String("credentials.hash", sh))
		account, err = c.authenticate(ctx, cookie.Value)
		if err != nil {
			return nil, connect.NewError(
				connect.CodeUnauthenticated,
				fmt.Errorf("authenticating by token: %w", err),
			)
		}
		log.Debug("user authenticated", zap.String("account.id", account.ID.String()))
		c.cache.Add(sh, account)
	}
	return account, nil
}

func (c *Controller) authenticate(ctx context.Context, credential string) (*db.Account, error) {
	info, err := idtoken.Validate(ctx, credential, c.cfg.Auth.Google.ClientID)
	if err != nil {
		return nil, fmt.Errorf("validating google credentials: %w", err)
	}

	claims := &GoogleClaims{}
	for k, v := range info.Claims {
		if k == "email" {
			claims.Email = v.(string)
		}
		if k == "email_verified" {
			claims.EmailVerified = v.(bool)
		}
		if k == "name" {
			claims.Name = v.(string)
		}
		if k == "picture" {
			claims.Picture = v.(string)
		}
	}
	if claims.Email == "" {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("email not detected"))
	}
	if !claims.EmailVerified {
		return nil, connect.NewError(connect.CodeFailedPrecondition, fmt.Errorf("email is not verified: %q", claims.Email))
	}

	var account db.Account
	err = c.cfg.Postgres.Tx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		account, err = q.GetAccount(ctx, claims.Email)
		if errors.Is(err, pgx.ErrNoRows) {
			account, err = q.CreateAccount(ctx, db.CreateAccountParams{
				Email:       claims.Email,
				DisplayName: claims.Email,
				Picture:     claims.Picture,
			})
			if err != nil {
				return fmt.Errorf("creating account: %w", err)
			}
		}
		return err
	})
	return &account, err
}
