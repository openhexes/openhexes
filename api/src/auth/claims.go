package auth

import "fmt"

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

func (c *Controller) setUpTestUsers() *Controller {
	if !c.cfg.Test.Enabled {
		return c
	}

	c.testUsers = map[string]*GoogleClaims{
		c.cfg.Test.Tokens.Owner: fillPicture(&GoogleClaims{
			Email:         "owner@test.com",
			EmailVerified: true,
			Name:          "Test Owner",
		}),
		c.cfg.Test.Tokens.Unverified: fillPicture(&GoogleClaims{
			Email: "unverified@test.com",
			Name:  "Test Unverified",
		}),
		c.cfg.Test.Tokens.Alfa: fillPicture(&GoogleClaims{
			Email:         "alfa@test.com",
			EmailVerified: true,
			Name:          "Test Alfa",
		}),
		c.cfg.Test.Tokens.Bravo: fillPicture(&GoogleClaims{
			Email:         "bravo@test.com",
			EmailVerified: true,
			Name:          "Test Bravo",
		}),
	}
	return c
}
