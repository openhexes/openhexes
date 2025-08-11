package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"slices"

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

	cookie, err := (&http.Request{Header: header}).Cookie(config.AuthCookieGoogle)
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
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("authenticating: %w", err))
		}
		log.Info("user authenticated", zap.String("account.id", account.ID.String()))
		c.cache.Add(sh, account)
	}
	return account, nil
}

func (c *Controller) resolveCredential(ctx context.Context, credential string) (*config.GoogleClaims, error) {
	log := config.GetLogger(ctx)

	if c.cfg.Test.Enabled {
		if claims, ok := c.cfg.Test.Accounts[credential]; ok {
			log.Info("resolved test account", zap.String("email", claims.Email))
			return claims, nil
		}
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid test account credentials"))
	}

	info, err := idtoken.Validate(ctx, credential, c.cfg.Auth.Google.ClientID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("validating google credentials: %w", err))
	}

	claims := &config.GoogleClaims{}
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
	return claims, nil
}

func (c *Controller) authenticate(ctx context.Context, credential string) (*db.Account, error) {
	claims, err := c.resolveCredential(ctx, credential)
	if err != nil {
		return nil, err
	}

	var account db.Account
	err = c.cfg.Postgres.Tx(ctx, func(tx pgx.Tx, q *db.Queries) error {
		account, err = q.GetAccountByEmail(ctx, claims.Email)
		if errors.Is(err, pgx.ErrNoRows) {
			var isOwner bool
			var isActive bool
			if slices.Contains(c.cfg.Auth.Owners.Emails, claims.Email) {
				isOwner = true
			}
			if slices.Contains(c.cfg.Auth.PreActivatedEmails, claims.Email) {
				isActive = true
			}

			account, err = q.CreateAccount(ctx, db.CreateAccountParams{
				Active:      isOwner || isActive,
				Email:       claims.Email,
				DisplayName: claims.Name,
				Picture:     claims.Picture,
			})
			if err != nil {
				return fmt.Errorf("creating account: %w", err)
			}

			if isOwner {
				for _, role := range c.cfg.Auth.Owners.Roles {
					err = q.GrantRole(ctx, db.GrantRoleParams{
						AccountID: account.ID,
						RoleID:    role,
					})
					if err != nil {
						return fmt.Errorf("granting role: %q -> %q: %w", role, claims.Email, err)
					}
				}
			}
		}
		return err
	})
	return &account, err
}
