package tiles

import (
	"fmt"
	"iter"

	mapv1 "github.com/openhexes/proto/map/v1"
)

type CoordinateKey struct {
	Depth  uint32
	Row    uint32
	Column uint32
}

func (k CoordinateKey) String() string {
	return fmt.Sprintf("%d.%d.%d", k.Depth, k.Row, k.Column)
}

type Index = map[CoordinateKey]*mapv1.Tile

type Neighbour struct {
	Direction     mapv1.EdgeDirection
	CoordinateKey CoordinateKey
}

func CoordinateToKey(c *mapv1.Tile_Coordinate) CoordinateKey {
	return CoordinateKey{Depth: c.Depth, Row: c.Row, Column: c.Column}
}

func EqualCoordinates(a, b *mapv1.Tile_Coordinate) bool {
	if a == nil && b == nil {
		return true
	} else if a == nil || b == nil {
		return false
	}
	return a.Row == b.Row && a.Column == b.Column && a.Depth == b.Depth
}

func IterNeighbours(c CoordinateKey) iter.Seq[Neighbour] {
	return func(yield func(Neighbour) bool) {
		var n Neighbour

		if c.Row%2 == 0 {
			if c.Column > 0 {
				n = Neighbour{
					Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_W,
					CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row, Column: c.Column - 1},
				}
				if !yield(n) {
					return
				}
			}

			if c.Row > 0 && c.Column > 0 {
				n = Neighbour{
					Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_NW,
					CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row - 1, Column: c.Column - 1},
				}
				if !yield(n) {
					return
				}
			}

			if c.Row > 0 {
				n = Neighbour{
					Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_NE,
					CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row - 1, Column: c.Column},
				}
				if !yield(n) {
					return
				}
			}

			n = Neighbour{
				Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_E,
				CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row, Column: c.Column + 1},
			}
			if !yield(n) {
				return
			}

			n = Neighbour{
				Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_SE,
				CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row + 1, Column: c.Column},
			}
			if !yield(n) {
				return
			}

			if c.Column > 0 {
				n = Neighbour{
					Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_SW,
					CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row + 1, Column: c.Column - 1},
				}
				if !yield(n) {
					return
				}
			}
		} else {
			if c.Column > 0 {
				n = Neighbour{
					Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_W,
					CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row, Column: c.Column - 1},
				}
				if !yield(n) {
					return
				}
			}

			if c.Row > 0 {
				n = Neighbour{
					Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_NW,
					CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row - 1, Column: c.Column},
				}
				if !yield(n) {
					return
				}
			}

			if c.Row > 0 {
				n = Neighbour{
					Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_NE,
					CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row - 1, Column: c.Column + 1},
				}
				if !yield(n) {
					return
				}
			}

			n = Neighbour{
				Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_E,
				CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row, Column: c.Column + 1},
			}
			if !yield(n) {
				return
			}

			n = Neighbour{
				Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_SE,
				CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row + 1, Column: c.Column + 1},
			}
			if !yield(n) {
				return
			}

			n = Neighbour{
				Direction:     mapv1.EdgeDirection_EDGE_DIRECTION_SW,
				CoordinateKey: CoordinateKey{Depth: c.Depth, Row: c.Row + 1, Column: c.Column},
			}
			if !yield(n) {
				return
			}
		}
	}
}
