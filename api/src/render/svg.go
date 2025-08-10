package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/tiles"
	mapv1 "github.com/openhexes/proto/map/v1"
)

const snapScale = 2000.0 // 1/2000 px grid
const wedgeRatio = 0.9

func snap(v float64) float64 {
	return math.Round(v*snapScale) / snapScale
}

func f64(v float64) string {
	return strconv.FormatFloat(snap(v), 'f', -1, 64)
}

// cssVarFill returns an SVG fill attribute string with a CSS variable
// name and a hardcoded fallback value (CSS variable syntax supports fallbacks).
func cssVarFill(varName, fallback string) string {
	return fmt.Sprintf(`fill="var(--%s, %s)"`, varName, fallback)
}

// terrainVarName converts a terrain ID (like "core/terrain/water") to a
// CSS variable name (like "terrain-core-terrain-water-fill").
func terrainVarName(prefix, terrainID string) string {
	safeID := strings.ToLower(strings.ReplaceAll(terrainID, "/", "-"))
	return fmt.Sprintf("%s-%s-fill", prefix, safeID)
}

// hexagonVerticesWorld returns the outer vertices of a pointy-top hexagon
// in world coordinates, given its row/column in the grid and the tile dimensions.
func hexagonVerticesWorld(row, column uint32) [6][2]float64 {
	isEvenRow := row%2 == 0
	xOrigin := float64(column) * config.TileWidth
	if !isEvenRow {
		xOrigin += float64(config.TileWidth) / 2
	}
	yOrigin := float64(row) * (float64(config.TileHeight) * 0.75)

	verticalStep := float64(config.TileHeight) / 4

	return [6][2]float64{
		{xOrigin + float64(config.TileWidth)/2, yOrigin + 0},                          // N
		{xOrigin + float64(config.TileWidth), yOrigin + verticalStep},                 // NE
		{xOrigin + float64(config.TileWidth), yOrigin + 3*verticalStep},               // SE
		{xOrigin + float64(config.TileWidth)/2, yOrigin + float64(config.TileHeight)}, // S
		{xOrigin + 0, yOrigin + 3*verticalStep},                                       // SW
		{xOrigin + 0, yOrigin + verticalStep},                                         // NW
	}
}

// insetHexagonVertices returns an inner hexagon scaled towards its center.
// scaleFactor < 1 moves vertices inward; scaleFactor = 1 means no change.
func insetHexagonVertices(outerVertices [6][2]float64, scaleFactor float64) [6][2]float64 {
	var centerX, centerY float64
	for _, point := range outerVertices {
		centerX += point[0]
		centerY += point[1]
	}
	centerX /= 6
	centerY /= 6

	var innerVertices [6][2]float64
	for index, point := range outerVertices {
		innerVertices[index][0] = centerX + (point[0]-centerX)*scaleFactor
		innerVertices[index][1] = centerY + (point[1]-centerY)*scaleFactor
	}
	return innerVertices
}

// wedgePath returns an SVG path "d" string for a wedge-shaped quad
// between the outer and inner hexagon along edge from vertex i to vertex j.
func wedgePath(outerVertices, innerVertices [6][2]float64, indexA, indexB int) string {
	outerA := outerVertices[indexA]
	outerB := outerVertices[indexB]
	innerA := innerVertices[indexA]
	innerB := innerVertices[indexB]

	return fmt.Sprintf(
		"M%g,%g L%g,%g L%g,%g L%g,%g Z",
		outerA[0], outerA[1],
		outerB[0], outerB[1],
		innerB[0], innerB[1],
		innerA[0], innerA[1],
	)
}

func equilateralThirdVertex(outerVertex, innerVertex [2]float64) (cx1, cy1, cx2, cy2 float64) {
	ax, ay := outerVertex[0], outerVertex[1] // A
	bx, by := innerVertex[0], innerVertex[1] // B

	dx, dy := bx-ax, by-ay

	cos60 := 0.5
	sin60 := math.Sqrt(3) / 2

	// Rotate +60° around A
	cx1 = ax + dx*cos60 - dy*sin60
	cy1 = ay + dx*sin60 + dy*cos60

	// Rotate -60° around A
	cx2 = ax + dx*cos60 + dy*sin60
	cy2 = ay - dx*sin60 + dy*cos60

	return
}

// cornerPath returns two SVG path "d" strings for the two triangles
// that form the corner wedge at vertex vertexIndex, using outer and inner hex vertices.
func cornerPath(outerVertices, innerVertices [6][2]float64, vertexIndex int) string {
	outerVertex := outerVertices[vertexIndex]
	innerVertex := innerVertices[vertexIndex]

	cx1, cy1, cx2, cy2 := equilateralThirdVertex(outerVertex, innerVertex)
	return fmt.Sprintf(
		"M%g,%g L%g,%g L%g,%g L%g,%g Z",
		outerVertex[0], outerVertex[1],
		cx1, cy1,
		innerVertex[0], innerVertex[1],
		cx2, cy2,
	)
}

// polygonPathData converts vertices into an SVG "d" path string.
func polygonPathData(vertices [6][2]float64) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("M%g,%g", vertices[0][0], vertices[0][1]))
	for i := 1; i < len(vertices); i++ {
		sb.WriteString(fmt.Sprintf(" L%g,%g", vertices[i][0], vertices[i][1]))
	}
	sb.WriteString(" Z")
	return sb.String()
}

// Given inclusive bounds [minRow..maxRow], [minColumn..maxColumn]
func segmentWorldRect(b *mapv1.Segment_Bounds) (minX, minY, width, height float64) {
	left := float64(b.MinColumn) * config.TileWidth
	if b.MinRow%2 != 0 {
		left += config.TileWidth / 2
	}

	right := float64(b.MaxColumn+1) * config.TileWidth
	if b.MaxRow%2 != 0 {
		right += config.TileWidth / 2
	}

	minX, width = left, right-left
	minY = float64(b.MinRow) * config.RowHeight
	bottom := float64(b.MaxRow)*config.RowHeight + config.TileHeight
	height = bottom - minY
	return
}

// hexagonPathD returns an SVG path for a canonical hex at (0,0) using config tile dimensions.
func hexagonPathD() string {
	v := config.TileHeight / 4.0
	return fmt.Sprintf(
		"M%g,%g L%g,%g L%g,%g L%g,%g L%g,%g L%g,%g Z",
		config.TileWidth/2.0, 0.0, // N
		config.TileWidth, v, // NE
		config.TileWidth, 3*v, // SE
		config.TileWidth/2.0, config.TileHeight, // S
		0.0, 3*v, // SW
		0.0, v, // NW
	)
}

// tileOriginWorld returns the world top-left origin for a tile (row/column).
func tileOriginWorld(row, column uint32) (float64, float64) {
	isEvenRow := row%2 == 0
	xOrigin := float64(column) * config.TileWidth
	if !isEvenRow {
		xOrigin += config.TileWidth / 2.0
	}
	yOrigin := float64(row) * config.RowHeight
	return xOrigin, yOrigin
}

// EdgeSegmentByDirection maps an EdgeDirection enum to two vertex indexes.
func EdgeSegmentByDirection(direction mapv1.EdgeDirection) [2]int {
	switch direction {
	case mapv1.EdgeDirection_EDGE_DIRECTION_W:
		return [2]int{4, 5}
	case mapv1.EdgeDirection_EDGE_DIRECTION_NW:
		return [2]int{5, 0}
	case mapv1.EdgeDirection_EDGE_DIRECTION_NE:
		return [2]int{0, 1}
	case mapv1.EdgeDirection_EDGE_DIRECTION_E:
		return [2]int{1, 2}
	case mapv1.EdgeDirection_EDGE_DIRECTION_SE:
		return [2]int{2, 3}
	case mapv1.EdgeDirection_EDGE_DIRECTION_SW:
		return [2]int{3, 4}
	default:
		return [2]int{0, 0}
	}
}

// CornerVertexIndex maps a CornerDirection enum to a single vertex index.
func CornerVertexIndex(direction mapv1.CornerDirection) int {
	switch direction {
	case mapv1.CornerDirection_CORNER_DIRECTION_NW:
		return 5
	case mapv1.CornerDirection_CORNER_DIRECTION_N:
		return 0
	case mapv1.CornerDirection_CORNER_DIRECTION_NE:
		return 1
	case mapv1.CornerDirection_CORNER_DIRECTION_SE:
		return 2
	case mapv1.CornerDirection_CORNER_DIRECTION_S:
		return 3
	case mapv1.CornerDirection_CORNER_DIRECTION_SW:
		return 4
	default:
		return 0
	}
}

func getSegmentKey(segment *mapv1.Segment) string {
	b := segment.Bounds
	return fmt.Sprintf("segment-%d-%d-%d-%d", b.MinRow, b.MaxRow, b.MinColumn, b.MaxColumn)
}

// generateTerrainPatterns creates SVG pattern definitions for all terrain types
func generateTerrainPatterns() string {
	var patterns strings.Builder

	// Water - complex wave pattern
	patterns.WriteString(`<pattern id="pattern-water" patternUnits="userSpaceOnUse" width="56" height="28">`)
	patterns.WriteString(`<path fill="var(--terrain-water-pattern, rgba(156,146,172,0.4))" d="M56 26v2h-7.75c2.3-1.27 4.94-2 7.75-2zm-26 2a2 2 0 1 0-4 0h-4.09A25.98 25.98 0 0 0 0 16v-2c.67 0 1.34.02 2 .07V14a2 2 0 0 0-2-2v-2a4 4 0 0 1 3.98 3.6 28.09 28.09 0 0 1 2.8-3.86A8 8 0 0 0 0 6V4a9.99 9.99 0 0 1 8.17 4.23c.94-.95 1.96-1.83 3.03-2.63A13.98 13.98 0 0 0 0 0h7.75c2 1.1 3.73 2.63 5.1 4.45 1.12-.72 2.3-1.37 3.53-1.93A20.1 20.1 0 0 0 14.28 0h2.7c.45.56.88 1.14 1.29 1.74 1.3-.48 2.63-.87 4-1.15-.11-.2-.23-.4-.36-.59H26v.07a28.4 28.4 0 0 1 4 0V0h4.09l-.37.59c1.38.28 2.72.67 4.01 1.15.4-.6.84-1.18 1.3-1.74h2.69a20.1 20.1 0 0 0-2.1 2.52c1.23.56 2.41 1.2 3.54 1.93A16.08 16.08 0 0 1 48.25 0H56c-4.58 0-8.65 2.2-11.2 5.6 1.07.8 2.09 1.68 3.03 2.63A9.99 9.99 0 0 1 56 4v2a8 8 0 0 0-6.77 3.74c1.03 1.2 1.97 2.5 2.79 3.86A4 4 0 0 1 56 10v2a2 2 0 0 0-2 2.07 28.4 28.4 0 0 1 2-.07v2c-9.2 0-17.3 4.78-21.91 12H30zM7.75 28H0v-2c2.81 0 5.46.73 7.75 2zM56 20v2c-5.6 0-10.65 2.3-14.28 6h-2.7c4.04-4.89 10.15-8 16.98-8zm-39.03 8h-2.69C10.65 24.3 5.6 22 0 22v-2c6.83 0 12.94 3.11 16.97 8zm15.01-.4a28.09 28.09 0 0 1 2.8-3.86 8 8 0 0 0-13.55 0c1.03 1.2 1.97 2.5 2.79 3.86a4 4 0 0 1 7.96 0zm14.29-11.86c1.3-.48 2.63-.87 4-1.15a25.99 25.99 0 0 0-44.55 0c1.38.28 2.72.67 4.01 1.15a21.98 21.98 0 0 1 36.54 0zm-5.43 2.71c1.13-.72 2.3-1.37 3.54-1.93a19.98 19.98 0 0 0-32.76 0c1.23.56 2.41 1.2 3.54 1.93a15.98 15.98 0 0 1 25.68 0zm-4.67 3.78c.94-.95 1.96-1.83 3.03-2.63a13.98 13.98 0 0 0-22.4 0c1.07.8 2.09 1.68 3.03 2.63a9.99 9.99 0 0 1 16.34 0z"/>`)
	patterns.WriteString(`</pattern>`)

	// Grass - geometric blade pattern
	patterns.WriteString(`<pattern id="pattern-grass" patternUnits="userSpaceOnUse" width="20" height="20">`)
	patterns.WriteString(`<g fill="var(--terrain-grass-pattern, rgba(156,146,172,0.4))">`)
	patterns.WriteString(`<path d="M3 17h1v3h-1v-3zm3-4h1v7h-1v-7zm3 2h1v5h-1v-5zm6-8h1v11h-1V7zm3 4h1v7h-1v-7z"/>`)
	patterns.WriteString(`</g></pattern>`)

	// Highlands - reusing subterranean pattern
	patterns.WriteString(`<pattern id="pattern-highlands" patternUnits="userSpaceOnUse" width="60" height="60">`)
	patterns.WriteString(`<path d="M54.627 0l.83.828-1.415 1.415L51.8 0h2.827zM5.373 0l-.83.828L5.96 2.243 8.2 0H5.374zM48.97 0l3.657 3.657-1.414 1.414L46.143 0h2.828zM11.03 0L7.372 3.657 8.787 5.07 13.857 0H11.03zm32.284 0L49.8 6.485 48.384 7.9l-7.9-7.9h2.83zM16.686 0L10.2 6.485 11.616 7.9l7.9-7.9h-2.83zm20.97 0l9.315 9.314-1.414 1.414L34.828 0h2.83zM22.344 0L13.03 9.314l1.414 1.414L25.172 0h-2.83zM32 0l12.142 12.142-1.414 1.414L30 .828 17.272 13.556l-1.414-1.414L28 0h4zM.284 0l28 28-1.414 1.414L0 2.544V0h.284zM0 5.373l25.456 25.455-1.414 1.415L0 8.2V5.374zm0 5.656l22.627 22.627-1.414 1.414L0 13.86v-2.83zm0 5.656l19.8 19.8-1.415 1.413L0 19.514v-2.83zm0 5.657l16.97 16.97-1.414 1.415L0 25.172v-2.83zM0 28l14.142 14.142-1.414 1.414L0 30.828V28zm0 5.657L11.314 44.97 9.9 46.386l-9.9-9.9v-2.828zm0 5.657L8.485 47.8 7.07 49.212 0 42.143v-2.83zm0 5.657l5.657 5.657-1.414 1.415L0 47.8v-2.83zm0 5.657l2.828 2.83-1.414 1.413L0 53.456v-2.83zM54.627 60L30 35.373 5.373 60H8.2L30 38.2 51.8 60h2.827zm-5.656 0L30 41.03 11.03 60h2.828L30 43.858 46.142 60h2.83zm-5.656 0L30 46.686 16.686 60h2.83L30 49.515 40.485 60h2.83zm-5.657 0L30 52.343 22.343 60h2.83L30 55.172 34.828 60h2.83zM32 60l-2-2-2 2h4zM59.716 0l-28 28 1.414 1.414L60 2.544V0h-.284zM60 5.373L34.544 30.828l1.414 1.415L60 8.2V5.374zm0 5.656L37.373 33.656l1.414 1.414L60 13.86v-2.83zm0 5.656l-19.8 19.8 1.415 1.413L60 19.514v-2.83zm0 5.657l-16.97 16.97 1.414 1.415L60 25.172v-2.83zM60 28L45.858 42.142l1.414 1.414L60 30.828V28zm0 5.657L48.686 44.97l1.415 1.415 9.9-9.9v-2.828zm0 5.657L51.515 47.8l1.414 1.413 7.07-7.07v-2.83zm0 5.657l-5.657 5.657 1.414 1.415L60 47.8v-2.83zm0 5.657l-2.828 2.83 1.414 1.413L60 53.456v-2.83zM39.9 16.385l1.414-1.414L30 3.658 18.686 14.97l1.415 1.415 9.9-9.9 9.9 9.9zm-2.83 2.828l1.415-1.414L30 9.313 21.515 17.8l1.414 1.413 7.07-7.07 7.07 7.07zm-2.827 2.83l1.414-1.416L30 14.97l-5.657 5.657 1.414 1.415L30 17.8l4.243 4.242zm-2.83 2.827l1.415-1.414L30 20.626l-2.828 2.83 1.414 1.414L30 23.456l1.414 1.414zM56.87 59.414L58.284 58 30 29.716 1.716 58l1.414 1.414L30 32.544l26.87 26.87z" fill="var(--terrain-highlands-pattern, rgba(156,146,172,0.4))" fill-rule="evenodd"/>`)
	patterns.WriteString(`</pattern>`)

	// Dirt - simple capsule pattern
	patterns.WriteString(`<pattern id="pattern-dirt" patternUnits="userSpaceOnUse" width="12" height="16">`)
	patterns.WriteString(`<path d="M4 .99C4 .445 4.444 0 5 0c.552 0 1 .45 1 .99v4.02C6 5.555 5.556 6 5 6c-.552 0-1-.45-1-.99V.99zm6 8c0-.546.444-.99 1-.99.552 0 1 .45 1 .99v4.02c0 .546-.444.99-1 .99-.552 0-1-.45-1-.99V8.99z" fill="var(--terrain-dirt-pattern, rgba(156,146,172,0.4))" fill-rule="evenodd"/>`)
	patterns.WriteString(`</pattern>`)

	// Ash - simple capsule pattern
	patterns.WriteString(`<pattern id="pattern-ash" patternUnits="userSpaceOnUse" width="12" height="16">`)
	patterns.WriteString(`<path d="M4 .99C4 .445 4.444 0 5 0c.552 0 1 .45 1 .99v4.02C6 5.555 5.556 6 5 6c-.552 0-1-.45-1-.99V.99zm6 8c0-.546.444-.99 1-.99.552 0 1 .45 1 .99v4.02c0 .546-.444.99-1 .99-.552 0-1-.45-1-.99V8.99z" fill="var(--terrain-ash-pattern, rgba(156,146,172,0.4))" fill-rule="evenodd"/>`)
	patterns.WriteString(`</pattern>`)

	// Subterranean - complex geometric pattern
	patterns.WriteString(`<pattern id="pattern-subterranean" patternUnits="userSpaceOnUse" width="60" height="60">`)
	patterns.WriteString(`<path d="M54.627 0l.83.828-1.415 1.415L51.8 0h2.827zM5.373 0l-.83.828L5.96 2.243 8.2 0H5.374zM48.97 0l3.657 3.657-1.414 1.414L46.143 0h2.828zM11.03 0L7.372 3.657 8.787 5.07 13.857 0H11.03zm32.284 0L49.8 6.485 48.384 7.9l-7.9-7.9h2.83zM16.686 0L10.2 6.485 11.616 7.9l7.9-7.9h-2.83zm20.97 0l9.315 9.314-1.414 1.414L34.828 0h2.83zM22.344 0L13.03 9.314l1.414 1.414L25.172 0h-2.83zM32 0l12.142 12.142-1.414 1.414L30 .828 17.272 13.556l-1.414-1.414L28 0h4zM.284 0l28 28-1.414 1.414L0 2.544V0h.284zM0 5.373l25.456 25.455-1.414 1.415L0 8.2V5.374zm0 5.656l22.627 22.627-1.414 1.414L0 13.86v-2.83zm0 5.656l19.8 19.8-1.415 1.413L0 19.514v-2.83zm0 5.657l16.97 16.97-1.414 1.415L0 25.172v-2.83zM0 28l14.142 14.142-1.414 1.414L0 30.828V28zm0 5.657L11.314 44.97 9.9 46.386l-9.9-9.9v-2.828zm0 5.657L8.485 47.8 7.07 49.212 0 42.143v-2.83zm0 5.657l5.657 5.657-1.414 1.415L0 47.8v-2.83zm0 5.657l2.828 2.83-1.414 1.413L0 53.456v-2.83zM54.627 60L30 35.373 5.373 60H8.2L30 38.2 51.8 60h2.827zm-5.656 0L30 41.03 11.03 60h2.828L30 43.858 46.142 60h2.83zm-5.656 0L30 46.686 16.686 60h2.83L30 49.515 40.485 60h2.83zm-5.657 0L30 52.343 22.343 60h2.83L30 55.172 34.828 60h2.83zM32 60l-2-2-2 2h4zM59.716 0l-28 28 1.414 1.414L60 2.544V0h-.284zM60 5.373L34.544 30.828l1.414 1.415L60 8.2V5.374zm0 5.656L37.373 33.656l1.414 1.414L60 13.86v-2.83zm0 5.656l-19.8 19.8 1.415 1.413L60 19.514v-2.83zm0 5.657l-16.97 16.97 1.414 1.415L60 25.172v-2.83zM60 28L45.858 42.142l1.414 1.414L60 30.828V28zm0 5.657L48.686 44.97l1.415 1.415 9.9-9.9v-2.828zm0 5.657L51.515 47.8l1.414 1.413 7.07-7.07v-2.83zm0 5.657l-5.657 5.657 1.414 1.415L60 47.8v-2.83zm0 5.657l-2.828 2.83 1.414 1.413L60 53.456v-2.83zM39.9 16.385l1.414-1.414L30 3.658 18.686 14.97l1.415 1.415 9.9-9.9 9.9 9.9zm-2.83 2.828l1.415-1.414L30 9.313 21.515 17.8l1.414 1.413 7.07-7.07 7.07 7.07zm-2.827 2.83l1.414-1.416L30 14.97l-5.657 5.657 1.414 1.415L30 17.8l4.243 4.242zm-2.83 2.827l1.415-1.414L30 20.626l-2.828 2.83 1.414 1.414L30 23.456l1.414 1.414zM56.87 59.414L58.284 58 30 29.716 1.716 58l1.414 1.414L30 32.544l26.87 26.87z" fill="var(--terrain-subterranean-pattern, rgba(156,146,172,0.4))" fill-rule="evenodd"/>`)
	patterns.WriteString(`</pattern>`)

	// Rough - reusing highlands pattern
	patterns.WriteString(`<pattern id="pattern-rough" patternUnits="userSpaceOnUse" width="16" height="32">`)
	patterns.WriteString(`<g fill="var(--terrain-rough-pattern, rgba(156,146,172,0.4))">`)
	patterns.WriteString(`<path fill-rule="evenodd" d="M0 24h4v2H0v-2zm0 4h6v2H0v-2zm0-8h2v2H0v-2zM0 0h4v2H0V0zm0 4h2v2H0V4zm16 20h-6v2h6v-2zm0 4H8v2h8v-2zm0-8h-4v2h4v-2zm0-20h-6v2h6V0zm0 4h-4v2h4V4zm-2 12h2v2h-2v-2zm0-8h2v2h-2V8zM2 8h10v2H2V8zm0 8h10v2H2v-2zm-2-4h14v2H0v-2zm4-8h6v2H4V4zm0 16h6v2H4v-2zM6 0h2v2H6V0zm0 24h2v2H6v-2z"/>`)
	patterns.WriteString(`</g></pattern>`)

	// Wasteland - simple capsule pattern
	patterns.WriteString(`<pattern id="pattern-wasteland" patternUnits="userSpaceOnUse" width="12" height="16">`)
	patterns.WriteString(`<path d="M4 .99C4 .445 4.444 0 5 0c.552 0 1 .45 1 .99v4.02C6 5.555 5.556 6 5 6c-.552 0-1-.45-1-.99V.99zm6 8c0-.546.444-.99 1-.99.552 0 1 .45 1 .99v4.02c0 .546-.444.99-1 .99-.552 0-1-.45-1-.99V8.99z" fill="var(--terrain-wasteland-pattern, rgba(156,146,172,0.4))" fill-rule="evenodd"/>`)
	patterns.WriteString(`</pattern>`)

	// Sand - reusing swamp pattern
	patterns.WriteString(`<pattern id="pattern-sand" patternUnits="userSpaceOnUse" width="52" height="26">`)
	patterns.WriteString(`<g fill="none" fill-rule="evenodd">`)
	patterns.WriteString(`<g fill="var(--terrain-sand-pattern, rgba(156,146,172,0.4))">`)
	patterns.WriteString(`<path d="M10 10c0-2.21-1.79-4-4-4-3.314 0-6-2.686-6-6h2c0 2.21 1.79 4 4 4 3.314 0 6 2.686 6 6 0 2.21 1.79 4 4 4 3.314 0 6 2.686 6 6 0 2.21 1.79 4 4 4v2c-3.314 0-6-2.686-6-6 0-2.21-1.79-4-4-4-3.314 0-6-2.686-6-6zm25.464-1.95l8.486 8.486-1.414 1.414-8.486-8.486 1.414-1.414z"/>`)
	patterns.WriteString(`</g></g></pattern>`)

	// Snow - reusing swamp pattern
	patterns.WriteString(`<pattern id="pattern-snow" patternUnits="userSpaceOnUse" width="52" height="26">`)
	patterns.WriteString(`<g fill="none" fill-rule="evenodd">`)
	patterns.WriteString(`<g fill="var(--terrain-snow-pattern, rgba(156,146,172,0.4))">`)
	patterns.WriteString(`<path d="M10 10c0-2.21-1.79-4-4-4-3.314 0-6-2.686-6-6h2c0 2.21 1.79 4 4 4 3.314 0 6 2.686 6 6 0 2.21 1.79 4 4 4 3.314 0 6 2.686 6 6 0 2.21 1.79 4 4 4v2c-3.314 0-6-2.686-6-6 0-2.21-1.79-4-4-4-3.314 0-6-2.686-6-6zm25.464-1.95l8.486 8.486-1.414 1.414-8.486-8.486 1.414-1.414z"/>`)
	patterns.WriteString(`</g></g></pattern>`)

	// Swamp - reusing swamp pattern
	patterns.WriteString(`<pattern id="pattern-swamp" patternUnits="userSpaceOnUse" width="52" height="26">`)
	patterns.WriteString(`<g fill="none" fill-rule="evenodd">`)
	patterns.WriteString(`<g fill="var(--terrain-swamp-pattern, rgba(156,146,172,0.4))">`)
	patterns.WriteString(`<path d="M10 10c0-2.21-1.79-4-4-4-3.314 0-6-2.686-6-6h2c0 2.21 1.79 4 4 4 3.314 0 6 2.686 6 6 0 2.21 1.79 4 4 4 3.314 0 6 2.686 6 6 0 2.21 1.79 4 4 4v2c-3.314 0-6-2.686-6-6 0-2.21-1.79-4-4-4-3.314 0-6-2.686-6-6zm25.464-1.95l8.486 8.486-1.414 1.414-8.486-8.486 1.414-1.414z"/>`)
	patterns.WriteString(`</g></g></pattern>`)

	// Abyss - void geometric pattern
	patterns.WriteString(`<pattern id="pattern-abyss" patternUnits="userSpaceOnUse" width="12" height="12">`)
	patterns.WriteString(`<g fill="var(--terrain-abyss-pattern, rgba(156,146,172,0.4))">`)
	patterns.WriteString(`<path d="M0 0h1v1H0V0zm2 2h1v1H2V2zm4 0h1v1H6V2zm2 2h1v1H8V4zm-6 2h1v1H2V6zm4 0h1v1H6V6zm2 2h1v1H8V8zm-6 2h1v1H2v-1zm4 0h1v1H6v-1z"/>`)
	patterns.WriteString(`</g></pattern>`)

	return patterns.String()
}

// getTerrainPatternURL returns the pattern URL for a given terrain ID
func getTerrainPatternURL(terrainID string) string {
	switch terrainID {
	case "water":
		return "url(#pattern-water)"
	case "grass":
		return "url(#pattern-grass)"
	case "highlands":
		return "url(#pattern-highlands)"
	case "dirt":
		return "url(#pattern-dirt)"
	case "ash":
		return "url(#pattern-ash)"
	case "subterranean":
		return "url(#pattern-subterranean)"
	case "rough":
		return "url(#pattern-rough)"
	case "wasteland":
		return "url(#pattern-wasteland)"
	case "sand":
		return "url(#pattern-sand)"
	case "snow":
		return "url(#pattern-snow)"
	case "swamp":
		return "url(#pattern-swamp)"
	case "abyss":
		return "url(#pattern-abyss)"
	default:
		return ""
	}
}

const (
	svgHeader = `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="%s %s %s %s" preserveAspectRatio="none">`
	// %s (1): hex symbol id, %s (2): hex path d, %s (3): clip id, %s (4): <use> elements, %s (5): patterns
	svgDefsHex = `<defs><path id="%s" d="%s"/><clipPath id="%s" clipPathUnits="userSpaceOnUse">%s</clipPath>%s</defs>`
)

func GenerateSVGSegment(segment *mapv1.Segment, tileIndex tiles.Index) string {
	var builder strings.Builder

	minX, minY, segW, segH := segmentWorldRect(segment.Bounds)
	key := getSegmentKey(segment)

	// SVG root
	fmt.Fprintf(&builder, svgHeader, f64(segW), f64(segH), f64(minX), f64(minY), f64(segW), f64(segH))

	// Build the clip path uses: one <use> for each tile that could be rendered in this segment
	// This includes tiles with 1-tile overlap to eliminate seams
	hexSymbolID := key + "-hex"
	clipID := key + "-clip"

	var uses strings.Builder
	for row := segment.Bounds.MinRow - 1; row <= segment.Bounds.MaxRow+1; row++ {
		for col := segment.Bounds.MinColumn - 1; col <= segment.Bounds.MaxColumn+1; col++ {
			ox, oy := tileOriginWorld(uint32(row), uint32(col))
			fmt.Fprintf(&uses, `<use href="#%s" x="%s" y="%s"/>`, hexSymbolID, f64(ox), f64(oy))
		}
	}

	// Emit defs: hex symbol + union clip made from <use> items + patterns
	fmt.Fprintf(&builder, svgDefsHex, hexSymbolID, hexagonPathD(), clipID, uses.String(), generateTerrainPatterns())

	// Everything we draw is clipped to the union of the hexes in this segment
	fmt.Fprintf(&builder, `<g clip-path="url(#%s)" shape-rendering="geometricPrecision">`, clipID)

	// Collect all tiles to render (primary tiles + overlapping neighbors for seamless rendering)
	tilesToRender := make([]*mapv1.Tile, 0, len(segment.Tiles))
	tilesToRender = append(tilesToRender, segment.Tiles...)

	// Add neighboring tiles if we have an index and they would affect the rendering
	if tileIndex != nil {
		for row := segment.Bounds.MinRow - 1; row <= segment.Bounds.MaxRow+1; row++ {
			for col := segment.Bounds.MinColumn - 1; col <= segment.Bounds.MaxColumn+1; col++ {
				coordKey := tiles.CoordinateKey{Depth: 0, Row: uint32(row), Column: uint32(col)}
				if tile, exists := tileIndex[coordKey]; exists {
					// Check if this tile is not already in our primary list
					isAlreadyIncluded := false
					for _, primaryTile := range segment.Tiles {
						if primaryTile.Coordinate.Row == uint32(row) && primaryTile.Coordinate.Column == uint32(col) {
							isAlreadyIncluded = true
							break
						}
					}
					if !isAlreadyIncluded {
						tilesToRender = append(tilesToRender, tile)
					}
				}
			}
		}
	}

	for _, tile := range tilesToRender {
		// 1. Terrain fill
		outerVertices := hexagonVerticesWorld(tile.Coordinate.Row, tile.Coordinate.Column)
		terrainPath := polygonPathData(outerVertices)
		fmt.Fprintf(
			&builder,
			`<path d="%s" %s/>`,
			terrainPath,
			cssVarFill(terrainVarName("terrain", tile.TerrainId), "#78716c"),
		)

		// 1.5. Terrain pattern overlay
		patternURL := getTerrainPatternURL(tile.TerrainId)
		if patternURL != "" {
			fmt.Fprintf(
				&builder,
				`<path d="%s" fill="%s"/>`,
				terrainPath,
				patternURL,
			)
		}

		// 2. Edges
		if len(tile.RenderingSpec.Edges) > 0 {
			innerVertices := insetHexagonVertices(outerVertices, wedgeRatio)
			for _, edge := range tile.RenderingSpec.Edges {
				segment := EdgeSegmentByDirection(edge.Direction) // helper mapping dir → vertex indexes
				fmt.Fprintf(
					&builder,
					`<path d="%s" %s/>`,
					wedgePath(outerVertices, innerVertices, segment[0], segment[1]),
					cssVarFill(terrainVarName("edge", edge.NeighbourTerrainId), "rgba(255, 255, 255, 0.1)"),
				)
			}
		}

		// 3. Corners
		if len(tile.RenderingSpec.Corners) > 0 {
			innerVertices := insetHexagonVertices(outerVertices, wedgeRatio)
			for _, corner := range tile.RenderingSpec.Corners {
				var terrain *mapv1.Terrain
				for _, edge := range corner.Edges {
					t, ok := config.TerrainRegistry[edge.NeighbourTerrainId]
					if !ok {
						continue
					}
					if terrain == nil || t.RenderingSpec.RenderingType > terrain.RenderingSpec.RenderingType {
						terrain = t
					}
				}
				if terrain == nil {
					continue
				}

				vertexIndex := CornerVertexIndex(corner.Direction) // helper mapping dir → vertex index
				fmt.Fprintf(
					&builder,
					`<path d="%s" %s/>`,
					cornerPath(outerVertices, innerVertices, vertexIndex),
					cssVarFill(terrainVarName("corner", terrain.Id), "rgba(255, 255, 255, 0.1)"),
				)
			}
		}
	}

	builder.WriteString(`</g></svg>`)

	return builder.String()
}
