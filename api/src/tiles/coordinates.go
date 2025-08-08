package tiles

import (
	"fmt"
	"iter"

	mapv1 "github.com/openhexes/proto/map/v1"
)

func CoordinateToString(c *mapv1.Tile_Coordinate) string {
	return fmt.Sprintf("%d.%d.%d", c.Depth, c.Row, c.Column)
}

func EqualCoordinates(a, b *mapv1.Tile_Coordinate) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	return a.Row == b.Row && a.Column == b.Column && a.Depth == b.Depth
}

func GetNeighbours(c *mapv1.Tile_Coordinate) iter.Seq[*mapv1.Tile_Coordinate] {
	if c.Row%2 == 0 {
		return func(yield func(*mapv1.Tile_Coordinate) bool) {
			if c.Column > 0 {
				if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row, Column: c.Column - 1}) {
					return
				}
			}
			if c.Row > 0 && c.Column > 0 {
				if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row - 1, Column: c.Column - 1}) {
					return
				}
			}
			if c.Row > 0 {
				if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row - 1, Column: c.Column}) {
					return
				}
			}
			if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row, Column: c.Column + 1}) {
				return
			}
			if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row + 1, Column: c.Column}) {
				return
			}
			if c.Column > 0 {
				if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row + 1, Column: c.Column - 1}) {
					return
				}
			}
		}
	}

	return func(yield func(*mapv1.Tile_Coordinate) bool) {
		if c.Column > 0 {
			if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row, Column: c.Column - 1}) {
				return
			}
		}
		if c.Row > 0 {
			if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row - 1, Column: c.Column}) {
				return
			}
		}
		if c.Row > 0 {
			if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row - 1, Column: c.Column + 1}) {
				return
			}
		}
		if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row, Column: c.Column + 1}) {
			return
		}
		if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row + 1, Column: c.Column + 1}) {
			return
		}
		if !yield(&mapv1.Tile_Coordinate{Depth: c.Depth, Row: c.Row + 1, Column: c.Column}) {
			return
		}
	}
}
