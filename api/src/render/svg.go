package render

import (
	_ "embed"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/tiles"
	mapv1 "github.com/openhexes/proto/map/v1"
)

//go:embed terrain-patterns.svg.tmpl
var terrainPatternsTemplate string

const snapScale = 2000.0 // 1/2000 px grid
const wedgeRatio = 0.9

var (
	terrainFillColors = map[mapv1.Terrain_RenderingType]string{
		mapv1.Terrain_RENDERING_TYPE_ABYSS:        "#000000",
		mapv1.Terrain_RENDERING_TYPE_WATER:        "#2e4255",
		mapv1.Terrain_RENDERING_TYPE_GRASS:        "#4f7253",
		mapv1.Terrain_RENDERING_TYPE_HIGHLANDS:    "#507747",
		mapv1.Terrain_RENDERING_TYPE_DIRT:         "#605a42",
		mapv1.Terrain_RENDERING_TYPE_ASH:          "#2e2e2e",
		mapv1.Terrain_RENDERING_TYPE_SUBTERRANEAN: "#0a0a0a",
		mapv1.Terrain_RENDERING_TYPE_ROUGH:        "#3d3333",
		mapv1.Terrain_RENDERING_TYPE_WASTELAND:    "#a78d45",
		mapv1.Terrain_RENDERING_TYPE_SAND:         "#c9ae73",
		mapv1.Terrain_RENDERING_TYPE_SNOW:         "#c6d2d4",
		mapv1.Terrain_RENDERING_TYPE_SWAMP:        "#468e6e",
	}

	terrainAccentColors = map[mapv1.Terrain_RenderingType]string{
		mapv1.Terrain_RENDERING_TYPE_ABYSS:        "#000000",
		mapv1.Terrain_RENDERING_TYPE_WATER:        "#324a62",
		mapv1.Terrain_RENDERING_TYPE_GRASS:        "#3f6543",
		mapv1.Terrain_RENDERING_TYPE_HIGHLANDS:    "#47684b",
		mapv1.Terrain_RENDERING_TYPE_DIRT:         "#55503c",
		mapv1.Terrain_RENDERING_TYPE_ASH:          "#272727",
		mapv1.Terrain_RENDERING_TYPE_SUBTERRANEAN: "#171717",
		mapv1.Terrain_RENDERING_TYPE_ROUGH:        "#2d2727",
		mapv1.Terrain_RENDERING_TYPE_WASTELAND:    "#9e8644",
		mapv1.Terrain_RENDERING_TYPE_SAND:         "#bca36d",
		mapv1.Terrain_RENDERING_TYPE_SNOW:         "#bec8ca",
		mapv1.Terrain_RENDERING_TYPE_SWAMP:        "#5d9f81",
	}
)

func snap(v float64) float64 {
	return math.Round(v*snapScale) / snapScale
}

func f64(v float64) string {
	return strconv.FormatFloat(snap(v), 'f', -1, 64)
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

// segmentWorldRect calculates the world rectangle for a segment given inclusive bounds [minRow..maxRow], [minColumn..maxColumn]
// Returns the minimum X/Y coordinates and the width/height of the segment in world space
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

// generateTerrainPatterns creates SVG pattern definitions using embedded template
func generateTerrainPatterns() string {
	result := terrainPatternsTemplate
	
	// Replace fill placeholders with actual colors using enum numbers to keep synced
	replacements := make(map[string]string)
	
	// Generate ${FILL_} placeholders using actual enum values
	for renderingType, fillColor := range terrainFillColors {
		placeholder := fmt.Sprintf("${FILL_%d}", int32(renderingType))
		replacements[placeholder] = fillColor
	}
	
	// Generate ${ACCENT_} placeholders using actual enum values
	for renderingType, accentColor := range terrainAccentColors {
		placeholder := fmt.Sprintf("${ACCENT_%d}", int32(renderingType))
		replacements[placeholder] = accentColor
	}
	
	for placeholder, color := range replacements {
		result = strings.ReplaceAll(result, placeholder, color)
	}
	
	return result
}

// getTerrainPatternURL returns the pattern URL for a given terrain ID
func getTerrainPatternURL(terrainID string) string {
	switch terrainID {
	case "water":
		return "url(#water-pattern)"
	case "grass":
		return "url(#grass-pattern)"
	case "highlands":
		return "url(#highlands-pattern)"
	case "sand":
		return "url(#desert-pattern)"
	case "swamp":
		return "url(#swamp-pattern)"
	case "rough":
		return "url(#rough-pattern)"
	case "subterranean":
		return "url(#subterranean-pattern)"
	case "ash":
		return "url(#lava-pattern)"
	case "dirt":
		return "url(#dirt-pattern)"
	case "wasteland":
		return "url(#wasteland-pattern)"
	case "snow":
		return "url(#water-pattern)" // Use water pattern for snow (dots)
	case "abyss":
		return "" // No pattern for abyss (pure black)
	default:
		return ""
	}
}

const (
	svgHeader = `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="%s %s %s %s" preserveAspectRatio="none">`
	// %s (1): hex symbol id, %s (2): hex path d, %s (3): clip id, %s (4): <use> elements, %s (5): patterns
	svgDefsHex = `<defs><path id="%s" d="%s"/><clipPath id="%s" clipPathUnits="userSpaceOnUse">%s</clipPath>%s</defs>`
)

func GenerateSVGSegment(segment *mapv1.Segment, tileIndex tiles.Index, layerDepth uint32) (string, error) {
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
	startRow := segment.Bounds.MinRow - 1
	if startRow < 0 {
		startRow = 0
	}
	startCol := segment.Bounds.MinColumn - 1
	if startCol < 0 {
		startCol = 0
	}

	for row := startRow; row <= segment.Bounds.MaxRow+1; row++ {
		for col := startCol; col <= segment.Bounds.MaxColumn+1; col++ {
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
		// Use the layer depth passed as parameter (works for both empty and populated segments)
		segmentDepth := layerDepth

		for row := segment.Bounds.MinRow - 1; row <= segment.Bounds.MaxRow+1; row++ {
			for col := segment.Bounds.MinColumn - 1; col <= segment.Bounds.MaxColumn+1; col++ {
				coordKey := tiles.CoordinateKey{Depth: segmentDepth, Row: uint32(row), Column: uint32(col)}
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
		terrain, ok := config.TerrainRegistry[tile.TerrainId]
		if !ok {
			return "", fmt.Errorf("tile %s is invalid: terrain not registered: %q", tile.Key, tile.TerrainId)
		}

		// 1. Terrain fill
		outerVertices := hexagonVerticesWorld(tile.Coordinate.Row, tile.Coordinate.Column)
		terrainPath := polygonPathData(outerVertices)
		fillColor := terrainFillColors[terrain.RenderingSpec.RenderingType]
		fmt.Fprintf(
			&builder,
			`<path d="%s" fill="%s"/>`,
			terrainPath,
			fillColor,
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
				accentColor := terrainAccentColors[terrain.RenderingSpec.RenderingType]
				fmt.Fprintf(
					&builder,
					`<path d="%s" fill="%s"/>`,
					wedgePath(outerVertices, innerVertices, segment[0], segment[1]),
					accentColor,
				)
			}
		}

		// 3. Corners
		if len(tile.RenderingSpec.Corners) > 0 {
			innerVertices := insetHexagonVertices(outerVertices, wedgeRatio)
			for _, corner := range tile.RenderingSpec.Corners {
				var tt *mapv1.Terrain
				for _, edge := range corner.Edges {
					t, ok := config.TerrainRegistry[edge.NeighbourTerrainId]
					if !ok {
						return "", fmt.Errorf(
							"tile %q: corner %s: edge %s: invalid neighbour terrain: %q",
							tile.Key, corner.Direction, edge.Direction, edge.NeighbourTerrainId,
						)
					}
					if tt == nil || t.RenderingSpec.RenderingType > tt.RenderingSpec.RenderingType {
						tt = t
					}
				}
				if tt == nil {
					return "", fmt.Errorf("tile %q: corner %s: failed to detect terrain", tile.Key, corner.Direction)
				}

				vertexIndex := CornerVertexIndex(corner.Direction) // helper mapping dir → vertex index
				accentColor := terrainAccentColors[tt.RenderingSpec.RenderingType]
				fmt.Fprintf(
					&builder,
					`<path d="%s" fill="%s"/>`,
					cornerPath(outerVertices, innerVertices, vertexIndex),
					accentColor,
				)
			}
		}
	}

	builder.WriteString(`</g></svg>`)

	return builder.String(), nil
}

func GenerateLightweightSVGSegment(segment *mapv1.Segment, tileIndex tiles.Index, layerDepth uint32) (string, error) {
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
	startRow := segment.Bounds.MinRow - 1
	if startRow < 0 {
		startRow = 0
	}
	startCol := segment.Bounds.MinColumn - 1
	if startCol < 0 {
		startCol = 0
	}

	for row := startRow; row <= segment.Bounds.MaxRow+1; row++ {
		for col := startCol; col <= segment.Bounds.MaxColumn+1; col++ {
			ox, oy := tileOriginWorld(uint32(row), uint32(col))
			fmt.Fprintf(&uses, `<use href="#%s" x="%s" y="%s"/>`, hexSymbolID, f64(ox), f64(oy))
		}
	}

	// Emit defs: hex symbol + union clip made from <use> items (no patterns for lightweight version)
	fmt.Fprintf(&builder, `<defs><path id="%s" d="%s"/><clipPath id="%s" clipPathUnits="userSpaceOnUse">%s</clipPath></defs>`, hexSymbolID, hexagonPathD(), clipID, uses.String())

	// Everything we draw is clipped to the union of the hexes in this segment
	fmt.Fprintf(&builder, `<g clip-path="url(#%s)" shape-rendering="geometricPrecision">`, clipID)

	// Collect all tiles to render (primary tiles + overlapping neighbors for seamless rendering)
	tilesToRender := make([]*mapv1.Tile, 0, len(segment.Tiles))
	tilesToRender = append(tilesToRender, segment.Tiles...)

	// Add neighboring tiles if we have an index and they would affect the rendering
	if tileIndex != nil {
		// Use the layer depth passed as parameter (works for both empty and populated segments)
		segmentDepth := layerDepth

		for row := segment.Bounds.MinRow - 1; row <= segment.Bounds.MaxRow+1; row++ {
			for col := segment.Bounds.MinColumn - 1; col <= segment.Bounds.MaxColumn+1; col++ {
				coordKey := tiles.CoordinateKey{Depth: segmentDepth, Row: uint32(row), Column: uint32(col)}
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
		terrain, ok := config.TerrainRegistry[tile.TerrainId]
		if !ok {
			return "", fmt.Errorf("tile %s is invalid: terrain not registered: %q", tile.Key, tile.TerrainId)
		}

		// Only render terrain fill - no patterns, edges, or corners for lightweight version
		outerVertices := hexagonVerticesWorld(tile.Coordinate.Row, tile.Coordinate.Column)
		terrainPath := polygonPathData(outerVertices)
		fmt.Fprintf(
			&builder,
			`<path d="%s" fill="%s"/>`,
			terrainPath,
			terrainFillColors[terrain.RenderingSpec.RenderingType],
		)
	}

	builder.WriteString(`</g></svg>`)

	return builder.String(), nil
}
