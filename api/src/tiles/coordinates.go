package tiles

import (
	"fmt"
	"iter"

	mapv1 "github.com/openhexes/proto/map/v1"
)

var (
	AllCornerDirections = []mapv1.CornerDirection{
		mapv1.CornerDirection_CORNER_DIRECTION_N,
		mapv1.CornerDirection_CORNER_DIRECTION_NE,
		mapv1.CornerDirection_CORNER_DIRECTION_SE,
		mapv1.CornerDirection_CORNER_DIRECTION_S,
		mapv1.CornerDirection_CORNER_DIRECTION_SW,
		mapv1.CornerDirection_CORNER_DIRECTION_NW,
	}
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

type CornerNeighbour struct {
	CoordinateKey CoordinateKey
	EdgeDirection mapv1.EdgeDirection // target's edge direction, which touches the neighbour
}

func GetCornerNeighbours(target CoordinateKey, cornerDirection mapv1.CornerDirection) []CornerNeighbour {
	if target.Row%2 == 0 {
		switch cornerDirection {
		case mapv1.CornerDirection_CORNER_DIRECTION_N:
			if target.Row <= 0 {
				return nil
			}
			if target.Column <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NE,
					},
				}
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NW,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NE,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_NE:
			if target.Row <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column + 1},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_E,
					},
				}
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NE,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_E,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_SE:
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_E,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SE,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_S:
			if target.Column <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SE,
					},
				}
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SE,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SW,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_SW:
			if target.Column <= 0 {
				return nil
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SW,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_W,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_NW:
			if target.Column <= 0 {
				if target.Row <= 0 {
					return nil
				}
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NW,
					},
				}
			}
			if target.Row <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column - 1},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_W,
					},
				}
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_W,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NW,
				},
			}
		}
	} else {
		switch cornerDirection {
		case mapv1.CornerDirection_CORNER_DIRECTION_N:
			if target.Row <= 0 {
				return nil
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NW,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NE,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_NE:
			if target.Row <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column + 1},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_E,
					},
				}
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NE,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_E,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_SE:
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_E,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SE,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_S:
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column + 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SE,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SW,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_SW:
			if target.Column <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SW,
					},
				}
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row + 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_SW,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_W,
				},
			}
		case mapv1.CornerDirection_CORNER_DIRECTION_NW:
			if target.Row <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column - 1},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_W,
					},
				}
			}
			if target.Column <= 0 {
				return []CornerNeighbour{
					{
						CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column},
						EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NW,
					},
				}
			}
			return []CornerNeighbour{
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row, Column: target.Column - 1},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_W,
				},
				{
					CoordinateKey: CoordinateKey{Depth: target.Depth, Row: target.Row - 1, Column: target.Column},
					EdgeDirection: mapv1.EdgeDirection_EDGE_DIRECTION_NW,
				},
			}
		}
	}

	return nil
}

// key for lookup
type cornerEdgeKey struct {
	corner mapv1.CornerDirection
	edge   mapv1.EdgeDirection
}

// For pointy-top hexes with your enums.
// Each entry maps: (currentCorner, currentEdge) -> (neighborCorner, neighborEdge).
var cornerEdgeToNeighbor = map[cornerEdgeKey]struct {
	otherCorner mapv1.CornerDirection
	otherEdge   mapv1.EdgeDirection
}{
	// Crossing the NW edge → neighbor’s SE edge (endpoints: N↔S, NW↔SE)
	{mapv1.CornerDirection_CORNER_DIRECTION_N, mapv1.EdgeDirection_EDGE_DIRECTION_NW}:  {mapv1.CornerDirection_CORNER_DIRECTION_S, mapv1.EdgeDirection_EDGE_DIRECTION_SE},
	{mapv1.CornerDirection_CORNER_DIRECTION_NW, mapv1.EdgeDirection_EDGE_DIRECTION_NW}: {mapv1.CornerDirection_CORNER_DIRECTION_SE, mapv1.EdgeDirection_EDGE_DIRECTION_SE},

	// Crossing the NE edge → neighbor’s SW edge (endpoints: N↔S, NE↔SW)
	{mapv1.CornerDirection_CORNER_DIRECTION_N, mapv1.EdgeDirection_EDGE_DIRECTION_NE}:  {mapv1.CornerDirection_CORNER_DIRECTION_S, mapv1.EdgeDirection_EDGE_DIRECTION_SW},
	{mapv1.CornerDirection_CORNER_DIRECTION_NE, mapv1.EdgeDirection_EDGE_DIRECTION_NE}: {mapv1.CornerDirection_CORNER_DIRECTION_SW, mapv1.EdgeDirection_EDGE_DIRECTION_SW},

	// Crossing the E edge → neighbor’s W edge (endpoints: NE↔NW, SE↔SW)
	{mapv1.CornerDirection_CORNER_DIRECTION_NE, mapv1.EdgeDirection_EDGE_DIRECTION_E}: {mapv1.CornerDirection_CORNER_DIRECTION_NW, mapv1.EdgeDirection_EDGE_DIRECTION_W},
	{mapv1.CornerDirection_CORNER_DIRECTION_SE, mapv1.EdgeDirection_EDGE_DIRECTION_E}: {mapv1.CornerDirection_CORNER_DIRECTION_SW, mapv1.EdgeDirection_EDGE_DIRECTION_W},

	// Crossing the SE edge → neighbor’s NW edge (endpoints: SE↔NW, S↔N)
	{mapv1.CornerDirection_CORNER_DIRECTION_SE, mapv1.EdgeDirection_EDGE_DIRECTION_SE}: {mapv1.CornerDirection_CORNER_DIRECTION_N, mapv1.EdgeDirection_EDGE_DIRECTION_NW},
	{mapv1.CornerDirection_CORNER_DIRECTION_S, mapv1.EdgeDirection_EDGE_DIRECTION_SE}:  {mapv1.CornerDirection_CORNER_DIRECTION_N, mapv1.EdgeDirection_EDGE_DIRECTION_NW},

	// Crossing the SW edge → neighbor’s NE edge (endpoints: S↔N, SW↔NE)
	{mapv1.CornerDirection_CORNER_DIRECTION_S, mapv1.EdgeDirection_EDGE_DIRECTION_SW}:  {mapv1.CornerDirection_CORNER_DIRECTION_N, mapv1.EdgeDirection_EDGE_DIRECTION_NE},
	{mapv1.CornerDirection_CORNER_DIRECTION_SW, mapv1.EdgeDirection_EDGE_DIRECTION_SW}: {mapv1.CornerDirection_CORNER_DIRECTION_NE, mapv1.EdgeDirection_EDGE_DIRECTION_NE},

	// Crossing the W edge → neighbor’s E edge (endpoints: SW↔SE, NW↔NE)
	{mapv1.CornerDirection_CORNER_DIRECTION_SW, mapv1.EdgeDirection_EDGE_DIRECTION_W}: {mapv1.CornerDirection_CORNER_DIRECTION_SE, mapv1.EdgeDirection_EDGE_DIRECTION_E},
	{mapv1.CornerDirection_CORNER_DIRECTION_NW, mapv1.EdgeDirection_EDGE_DIRECTION_W}: {mapv1.CornerDirection_CORNER_DIRECTION_NE, mapv1.EdgeDirection_EDGE_DIRECTION_E},
}

func GetIntersectionOnCorner(
	corner mapv1.CornerDirection,
	edge mapv1.EdgeDirection,
) (otherCorner mapv1.CornerDirection, otherEdge mapv1.EdgeDirection) {
	if m, ok := cornerEdgeToNeighbor[cornerEdgeKey{corner, edge}]; ok {
		return m.otherCorner, m.otherEdge
	}
	// Not an incident edge for that corner, or unsupported input.
	return 0, 0
}
