package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openhexes/openhexes/api/src/config"
	mapv1 "github.com/openhexes/proto/map/v1"
)

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

// wedgePathData returns an SVG path "d" string for a wedge-shaped quad
// between the outer and inner hexagon along edge from vertex i to vertex j.
func wedgePathData(outerVertices, innerVertices [6][2]float64, indexA, indexB int) string {
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

// cornerTrianglesPathData returns two SVG path "d" strings for the two triangles
// that form the corner wedge at vertex vertexIndex, using outer and inner hex vertices.
func cornerTrianglesPathData(outerVertices, innerVertices [6][2]float64, vertexIndex int) (string, string) {
	previousIndex := (vertexIndex + 5) % 6
	nextIndex := (vertexIndex + 1) % 6

	outerVertex := outerVertices[vertexIndex]
	innerVertex := innerVertices[vertexIndex]
	innerPrev := innerVertices[previousIndex]
	innerNext := innerVertices[nextIndex]

	// Triangle 1: outer vertex → inner vertex → inner previous vertex
	pathData1 := fmt.Sprintf(
		"M%g,%g L%g,%g L%g,%g Z",
		outerVertex[0], outerVertex[1],
		innerVertex[0], innerVertex[1],
		innerPrev[0], innerPrev[1],
	)

	// Triangle 2: outer vertex → inner vertex → inner next vertex
	pathData2 := fmt.Sprintf(
		"M%g,%g L%g,%g L%g,%g Z",
		outerVertex[0], outerVertex[1],
		innerVertex[0], innerVertex[1],
		innerNext[0], innerNext[1],
	)

	return pathData1, pathData2
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

func f64(v float64) string { return strconv.FormatFloat(v, 'f', -1, 64) }

func getSegmentKey(segment *mapv1.Segment) string {
	b := segment.Bounds
	return fmt.Sprintf("segment-%d-%d-%d-%d", b.MinRow, b.MaxRow, b.MinColumn, b.MaxColumn)
}

const (
	svgHeader = `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="%s %s %s %s" preserveAspectRatio="none">`
	svgDefs   = `<defs><clipPath id="%s" clipPathUnits="userSpaceOnUse"><rect x="%s" y="%s" width="%s" height="%s"/></clipPath></defs>`
)

// GenerateSVGSegment returns an SVG string for all tiles in a segment.
func GenerateSVGSegment(segment *mapv1.Segment) string {
	var builder strings.Builder

	minX, minY, segW, segH := segmentWorldRect(segment.Bounds)
	key := getSegmentKey(segment)

	fmt.Fprintf(
		&builder, svgHeader,
		f64(segW), f64(segH), f64(minX), f64(minY), f64(segW), f64(segH),
	)

	fmt.Fprintf(
		&builder, svgDefs,
		key, f64(minX), f64(minY), f64(segW), f64(segH),
	)

	// everything gets clipped to the segment rect
	fmt.Fprintf(&builder, `<g clip-path="url(#%s)">`, key)

	for _, tile := range segment.Tiles {
		// 1. Terrain fill
		outerVertices := hexagonVerticesWorld(tile.Coordinate.Row, tile.Coordinate.Column)
		terrainPath := polygonPathData(outerVertices)
		fmt.Fprintf(
			&builder,
			`<path d="%s" %s/>`,
			terrainPath,
			cssVarFill(terrainVarName("terrain", tile.TerrainId), "#78716c"),
		)

		// 2. Edges
		if len(tile.RenderingSpec.Edges) > 0 {
			innerVertices := insetHexagonVertices(outerVertices, 0.9)
			for _, edge := range tile.RenderingSpec.Edges {
				segment := EdgeSegmentByDirection(edge.Direction) // helper mapping dir → vertex indexes
				wedge := wedgePathData(outerVertices, innerVertices, segment[0], segment[1])
				fmt.Fprintf(
					&builder,
					`<path d="%s" %s/>`,
					wedge,
					cssVarFill(terrainVarName("edge", edge.NeighbourTerrainId), "#44403c"),
				)
			}
		}

		// 3. Corners
		if len(tile.RenderingSpec.Corners) > 0 {
			innerVertices := insetHexagonVertices(outerVertices, 0.9)
			for _, corner := range tile.RenderingSpec.Corners {
				vertexIndex := CornerVertexIndex(corner.Direction) // helper mapping dir → vertex index
				triangle1, triangle2 := cornerTrianglesPathData(outerVertices, innerVertices, vertexIndex)

				// Neighbor 1
				if len(corner.NeighbourTerrainIds) > 0 {
					fmt.Fprintf(
						&builder,
						`<path d="%s" %s/>`,
						triangle1,
						cssVarFill(terrainVarName("corner", corner.NeighbourTerrainIds[0]), "#44403c"),
					)
				}
				// Neighbor 2
				if len(corner.NeighbourTerrainIds) > 1 {
					fmt.Fprintf(
						&builder,
						`<path d="%s" %s/>`,
						triangle2,
						cssVarFill(terrainVarName("corner", corner.NeighbourTerrainIds[1]), "#44403c"),
					)
				}
			}
		}
	}

	builder.WriteString(`</g></svg>`)

	return builder.String()
}
