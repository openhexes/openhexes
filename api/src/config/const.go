package config

import (
	"fmt"
	"math"
	"net/http"
)

const (
	AuthCookieGoogle = "hexes.auth.google"

	RoleOwner = "owner"

	// TileWidth and TileHeight are both in "world units" (same system used for SVG viewBox).
	TileHeight = float64(50)
)

var (
	CoreRoles = []string{
		RoleOwner,
	}

	TileWidth = TileHeight * math.Sqrt(3) / 2
	RowHeight = 0.75 * TileHeight
)

type RequestOption func(http.Header)

func WithToken(token string) RequestOption {
	return func(h http.Header) {
		h.Add("Cookie", fmt.Sprintf("%s=%s", AuthCookieGoogle, token))
	}
}
