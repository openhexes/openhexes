package config

import "math"

const (
	// TileWidth and TileHeight are both in "world units" (same system used for SVG viewBox).
	TileHeight = float64(60)
)

var (
	TileWidth = TileHeight * math.Sqrt(3) / 2
	RowHeight = 0.75 * TileHeight
)
