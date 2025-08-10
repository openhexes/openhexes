package game

import (
	"context"
	"fmt"
	"time"

	"connectrpc.com/connect"
	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/render"
	"github.com/openhexes/openhexes/api/src/server/progress"
	"github.com/openhexes/openhexes/api/src/tiles"
	gamev1 "github.com/openhexes/proto/game/v1"
	"github.com/openhexes/proto/game/v1/gamev1connect"
	mapv1 "github.com/openhexes/proto/map/v1"
	progressv1 "github.com/openhexes/proto/progress/v1"
	worldv1 "github.com/openhexes/proto/world/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

type Service struct {
	gamev1connect.UnimplementedGameServiceHandler

	cfg  *config.Config
	auth *auth.Controller
}

func New(cfg *config.Config, auth *auth.Controller) *Service {
	return &Service{
		cfg:  cfg,
		auth: auth,
	}
}

func (svc *Service) GetSampleWorld(ctx context.Context, request *connect.Request[gamev1.GetSampleWorldRequest], stream *connect.ServerStream[gamev1.GetSampleWorldResponse]) error {
	const (
		defaultTotalRows            = uint32(64)
		defaultTotalColumns         = uint32(64)
		defaultMaxRowsPerSegment    = uint32(15)
		defaultMaxColumnsPerSegment = uint32(15)
		defaultMaxChunkSizeBytes    = 32 * 1024 // 32Kb
	)

	if request.Msg.TotalRows == uint32(0) {
		request.Msg.TotalRows = defaultTotalRows
	}
	if request.Msg.TotalColumns == uint32(0) {
		request.Msg.TotalColumns = defaultTotalColumns
	}
	if request.Msg.MaxRowsPerSegment == uint32(0) {
		request.Msg.MaxRowsPerSegment = defaultMaxRowsPerSegment
	}
	if request.Msg.MaxColumnsPerSegment == uint32(0) {
		request.Msg.MaxColumnsPerSegment = defaultMaxColumnsPerSegment
	}

	stageGrid := &progressv1.Stage{
		State: progressv1.Stage_STATE_RUNNING,
		Title: "Prepare grid",
	}
	stageTiles := &progressv1.Stage{
		State: progressv1.Stage_STATE_WAITING,
		Title: "Process tiles",
	}
	stageEdges := &progressv1.Stage{
		State: progressv1.Stage_STATE_WAITING,
		Title: "Process edges",
	}
	stageCorners := &progressv1.Stage{
		State: progressv1.Stage_STATE_WAITING,
		Title: "Process corners",
	}
	stageRender := &progressv1.Stage{
		State: progressv1.Stage_STATE_WAITING,
		Title: "Render segments",
	}
	reporter := progress.NewReporter(
		ctx,
		func(p *progressv1.Progress) error {
			return stream.Send(&gamev1.GetSampleWorldResponse{
				Progress: p,
			})
		},
		stageGrid, stageTiles, stageEdges, stageCorners, stageRender,
	)
	defer reporter.Close()
	reporter.Update()

	// prepare segment containers
	start := time.Now()
	segmentsLength := request.Msg.TotalRows / request.Msg.MaxRowsPerSegment * request.Msg.TotalColumns / request.Msg.MaxColumnsPerSegment
	segments := make([]*mapv1.Segment, 0, segmentsLength)

	for rowStart := uint32(0); rowStart < request.Msg.TotalRows; rowStart += request.Msg.MaxRowsPerSegment {
		for columnStart := uint32(0); columnStart < request.Msg.TotalColumns; columnStart += request.Msg.MaxColumnsPerSegment {
			rowEnd := rowStart + request.Msg.MaxRowsPerSegment
			columnEnd := columnStart + request.Msg.MaxColumnsPerSegment
			segments = append(segments, &mapv1.Segment{
				Tiles: make([]*mapv1.Tile, 0, request.Msg.MaxRowsPerSegment*request.Msg.MaxColumnsPerSegment),
				Bounds: &mapv1.Segment_Bounds{
					MinRow:    int32(rowStart),
					MaxRow:    int32(rowEnd),
					MinColumn: int32(columnStart),
					MaxColumn: int32(columnEnd),
				},
			})
		}
	}

	// arrange segments in a grid
	gridRowLength := request.Msg.TotalRows / request.Msg.MaxRowsPerSegment
	segmentsPerRow := request.Msg.TotalColumns / request.Msg.MaxColumnsPerSegment
	segmentRows := make([]*mapv1.Segment_Row, 0, request.Msg.TotalRows/request.Msg.MaxRowsPerSegment)
	gridRow := make([]*mapv1.Segment, 0, segmentsPerRow)
	var rowStart *int32

	for _, segment := range segments {
		if rowStart == nil {
			rowStart = &segment.Bounds.MinRow
			gridRow = make([]*mapv1.Segment, 0, gridRowLength)
			gridRow = append(gridRow, segment)
		} else if *rowStart != segment.Bounds.MinRow {
			segmentRows = append(segmentRows, &mapv1.Segment_Row{Segments: gridRow})
			gridRow = make([]*mapv1.Segment, 0, gridRowLength)
			gridRow = append(gridRow, segment)
			rowStart = &segment.Bounds.MinRow
		} else {
			gridRow = append(gridRow, segment)
		}
	}
	segmentRows = append(segmentRows, &mapv1.Segment_Row{Segments: gridRow})

	stageGrid.Duration = durationpb.New(time.Since(start))
	stageGrid.State = progressv1.Stage_STATE_DONE
	stageTiles.State = progressv1.Stage_STATE_RUNNING
	reporter.Update()

	// generate tiles & put them into respective segments
	start = time.Now()
	totalTiles := request.Msg.TotalRows * request.Msg.TotalColumns
	var processedTileCount int

	islandCenters := []*mapv1.Tile_Coordinate{
		{
			Row:    0,
			Column: 0,
		},
		{
			Row:    8,
			Column: 8,
		},
		{
			Row:    10,
			Column: 10,
		},
		{
			Row:    11,
			Column: 7,
		},
		{
			Row:    12,
			Column: 12,
		},
		{
			Row:    16,
			Column: 10,
		},
	}

	islandSet := make(map[tiles.CoordinateKey]bool)
	for _, center := range islandCenters {
		ck := tiles.CoordinateToKey(center)

		islandSet[ck] = true
		for c := range tiles.IterNeighbours(ck) {
			islandSet[c.CoordinateKey] = true

			for cc := range tiles.IterNeighbours(c.CoordinateKey) {
				islandSet[cc.CoordinateKey] = true
			}
		}
	}

	idx := make(tiles.Index, totalTiles)
	for row := range request.Msg.TotalRows {
		for column := range request.Msg.TotalColumns {
			tile := &mapv1.Tile{
				Coordinate: &mapv1.Tile_Coordinate{
					Row:    uint32(row),
					Column: uint32(column),
				},
			}
			k := tiles.CoordinateToKey(tile.Coordinate)
			idx[k] = tile

			if islandSet[k] {
				tile.TerrainId = "ash"
			} else {
				tile.TerrainId = "water"
			}

			processedTileCount++
			if processedTileCount%10_000 == 0 {
				stageTiles.Subtitle = fmt.Sprintf("%d / %d", processedTileCount, totalTiles)
				reporter.Update(float64(processedTileCount) / float64(totalTiles))
			}
		}
	}

	// Assign tiles to their primary segments (no overlap for tile data)
	for row := range request.Msg.TotalRows {
		segRowIdx := row / request.Msg.MaxRowsPerSegment
		segRow := segmentRows[segRowIdx]

		for column := range request.Msg.TotalColumns {
			segColIdx := column / request.Msg.MaxColumnsPerSegment
			segment := segRow.Segments[segColIdx]

			coordinate := &mapv1.Tile_Coordinate{
				Row:    uint32(row),
				Column: uint32(column),
			}
			k := tiles.CoordinateToKey(coordinate)
			tile := idx[k]

			segment.Tiles = append(segment.Tiles, tile)
		}
	}

	stageTiles.Subtitle = fmt.Sprintf("%d", totalTiles)
	stageTiles.Duration = durationpb.New(time.Since(start))
	stageTiles.State = progressv1.Stage_STATE_DONE
	stageEdges.State = progressv1.Stage_STATE_RUNNING
	reporter.Update(0)

	// calculate edges
	start = time.Now()
	processedTileCount = 0

	for k, tile := range idx {
		tile.RenderingSpec = &mapv1.Tile_RenderingSpec{
			Edges:   make([]*mapv1.Tile_Edge, 0, 6),
			Corners: make([]*mapv1.Tile_Corner, 0, 6),
		}
		tileTerrain, ok := config.TerrainRegistry[tile.TerrainId]
		if !ok {
			return fmt.Errorf("unregistered terrain id: %s: %q", k, tile.TerrainId)
		}
		tileTerrainZ := tileTerrain.RenderingSpec.RenderingType.Number()

		for c := range tiles.IterNeighbours(k) {
			neighbour, ok := idx[c.CoordinateKey]
			if !ok || neighbour.TerrainId == tile.TerrainId {
				continue
			}

			neighbourTerrain, ok := config.TerrainRegistry[neighbour.TerrainId]
			if !ok {
				return fmt.Errorf("unregistered terrain id: %s (neighbour of %s): %q", c.CoordinateKey, k, neighbour.TerrainId)
			}
			neighbourTerrainZ := neighbourTerrain.RenderingSpec.RenderingType.Number()

			if tileTerrainZ > neighbourTerrainZ {
				continue
			}

			tile.RenderingSpec.Edges = append(tile.RenderingSpec.Edges, &mapv1.Tile_Edge{
				Direction:          c.Direction,
				NeighbourTerrainId: neighbour.TerrainId,
			})
		}

		processedTileCount++
		if processedTileCount%10_000 == 0 {
			stageEdges.Subtitle = fmt.Sprintf("%d / %d", processedTileCount, totalTiles)
			reporter.Update(float64(processedTileCount) / float64(totalTiles))
		}
	}

	stageEdges.Subtitle = fmt.Sprintf("%d", totalTiles)
	stageEdges.Duration = durationpb.New(time.Since(start))
	stageEdges.State = progressv1.Stage_STATE_DONE

	// calculate corners
	stageCorners.State = progressv1.Stage_STATE_RUNNING
	reporter.Update(0)

	start = time.Now()
	processedTileCount = 0

	for k := range idx {
		for _, cd := range tiles.AllCornerDirections {
			cornerNeighbours := tiles.GetCornerNeighbours(k, cd)
			
			// Handle cases with 1 or 2 neighbors (was previously only handling exactly 2)
			if len(cornerNeighbours) == 0 {
				continue
			}

			var validNeighbors []struct {
				cn tiles.CornerNeighbour
				n  *mapv1.Tile
				matchedCorner mapv1.CornerDirection
				matchedEdge   mapv1.EdgeDirection
				edge          *mapv1.Tile_Edge
			}

			// Process each neighbor
			for _, cn := range cornerNeighbours {
				n, ok := idx[cn.CoordinateKey]
				if !ok {
					continue
				}
				
				matchedCorner, matchedEdge := tiles.GetIntersectionOnCorner(cd, cn.EdgeDirection)
				var edge *mapv1.Tile_Edge
				for _, e := range n.RenderingSpec.Edges {
					if e.Direction == matchedEdge {
						edge = e
						break
					}
				}
				if edge == nil {
					continue
				}

				validNeighbors = append(validNeighbors, struct {
					cn tiles.CornerNeighbour
					n  *mapv1.Tile
					matchedCorner mapv1.CornerDirection
					matchedEdge   mapv1.EdgeDirection
					edge          *mapv1.Tile_Edge
				}{cn, n, matchedCorner, matchedEdge, edge})
			}

			// Create corner markers for all valid neighbors
			if len(validNeighbors) > 0 {
				fmt.Printf("\n\n[CORNER FOUND] @ %d.%d %s (%d neighbors)\n", k.Row, k.Column, cd, len(validNeighbors))
				
				// Add corner markers to neighbor tiles
				for _, vn := range validNeighbors {
					fmt.Printf("\t -> %d.%d %s (on %s)\n", vn.n.Coordinate.Row, vn.n.Coordinate.Column, vn.edge.Direction, vn.cn.EdgeDirection)
					fmt.Printf("\t\t -> FILL %s\n", vn.matchedCorner)

					vn.n.RenderingSpec.Corners = append(vn.n.RenderingSpec.Corners, &mapv1.Tile_Corner{
						Direction: vn.matchedCorner,
						Edge:      vn.edge,
					})
				}
			}
		}

		processedTileCount++
		if processedTileCount%10_000 == 0 {
			stageCorners.Subtitle = fmt.Sprintf("%d / %d", processedTileCount, totalTiles)
			reporter.Update(float64(processedTileCount) / float64(totalTiles))
		}
	}

	stageCorners.Subtitle = fmt.Sprintf("%d", totalTiles)
	stageCorners.Duration = durationpb.New(time.Since(start))
	stageCorners.State = progressv1.Stage_STATE_DONE
	stageRender.State = progressv1.Stage_STATE_RUNNING
	reporter.Update(0)

	start = time.Now()
	var processedSegmentCount int

	for _, row := range segmentRows {
		for _, segment := range row.Segments {
			segment.RenderingSpec = &mapv1.Segment_RenderingSpec{
				Svg: render.GenerateSVGSegment(segment, idx),
			}

			processedSegmentCount++
			if processedSegmentCount%100 == 0 {
				stageRender.Subtitle = fmt.Sprintf("%d / %d", processedSegmentCount, len(segments))
				reporter.Update(float64(processedSegmentCount) / float64(len(segments)))
			}
		}
	}

	stageRender.Subtitle = fmt.Sprintf("%d", len(segments))
	stageRender.Duration = durationpb.New(time.Since(start))
	stageRender.State = progressv1.Stage_STATE_DONE
	reporter.Update(1)

	// generation done, time to send

	// grid dimensions
	dimensionsResponse := &gamev1.GetSampleWorldResponse{
		World: &worldv1.World{
			RenderingSpec: &worldv1.World_RenderingSpec{
				TileHeight: config.TileHeight,
				TileWidth:  config.TileWidth,
			},
			Layers: []*mapv1.Grid{
				{
					TotalRows:    request.Msg.TotalRows,
					TotalColumns: request.Msg.TotalColumns,
				},
			},
		},
	}
	if err := stream.Send(dimensionsResponse); err != nil {
		return err
	}

	// registries
	terrainsResponse := &gamev1.GetSampleWorldResponse{
		World: &worldv1.World{
			TerrainRegistry: make(map[string]*mapv1.Terrain),
		},
	}
	for key := range config.TerrainRegistry {
		terrainsResponse.World.TerrainRegistry[key] = config.TerrainRegistry[key]
		if proto.Size(terrainsResponse) >= defaultMaxChunkSizeBytes {
			if err := stream.Send(terrainsResponse); err != nil {
				return err
			}
			terrainsResponse = &gamev1.GetSampleWorldResponse{
				World: &worldv1.World{
					TerrainRegistry: make(map[string]*mapv1.Terrain),
				},
			}
		}
	}

	// tiles
	grid := &mapv1.Grid{
		SegmentRows: make([]*mapv1.Segment_Row, 0, 100),
	}
	for _, row := range segmentRows {
		grid.SegmentRows = append(grid.SegmentRows, row)
		if proto.Size(grid) >= defaultMaxChunkSizeBytes {
			response := &gamev1.GetSampleWorldResponse{
				World: &worldv1.World{
					Layers: []*mapv1.Grid{grid},
				},
			}
			if err := stream.Send(response); err != nil {
				return err
			}
			grid = &mapv1.Grid{
				SegmentRows: make([]*mapv1.Segment_Row, 0, 100),
			}
		}
	}
	if len(grid.SegmentRows) > 0 {
		response := &gamev1.GetSampleWorldResponse{
			World: &worldv1.World{
				Layers: []*mapv1.Grid{grid},
			},
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}

	return nil
}

func BoundsInclude(b *mapv1.Segment_Bounds, t *mapv1.Tile, modifier int32) bool {
	c := t.GetCoordinate()
	row := int32(c.GetRow())
	column := int32(c.GetColumn())

	return row >= (b.GetMinRow()-modifier) &&
		row < (b.GetMaxRow()+modifier) &&
		column >= (b.GetMinColumn()-modifier) &&
		column < (b.GetMaxColumn()+modifier)
}

// // shouldCreateCorner determines if a corner fill is needed at this vertex
// // This happens when 3+ different terrains meet at a vertex but there are fewer than 3 edges
// func shouldCreateCorner(tileKey tiles.CoordinateKey, cornerDir mapv1.CornerDirection, tileIndex tiles.Index) bool {
// 	currentTile, exists := tileIndex[tileKey]
// 	if !exists {
// 		return false
// 	}

// 	// Get all terrains that meet at this vertex (including current tile)
// 	terrainsAtVertex := getUniqueTerrainsAtVertex(tileKey, cornerDir, tileIndex)

// 	// Count how many edges this tile has that touch this vertex
// 	edgesAtVertex := countEdgesAtVertex(currentTile, cornerDir)

// 	// Need corner fill if 3+ terrains meet but fewer than 3 edges exist
// 	return len(terrainsAtVertex) >= 3 && edgesAtVertex < 3
// }

// // getCornerTerrains returns the terrain IDs for neighbors that need corner fills
// func getCornerTerrains(tileKey tiles.CoordinateKey, cornerDir mapv1.CornerDirection, tileIndex tiles.Index) []string {
// 	// Get the two neighbor tiles that share this vertex
// 	neighbors := getVertexNeighbors(tileKey, cornerDir)

// 	var terrains []string
// 	for _, neighborKey := range neighbors {
// 		if neighbor, exists := tileIndex[neighborKey]; exists {
// 			terrains = append(terrains, neighbor.TerrainId)
// 		}
// 	}

// 	return terrains
// }

// // getUniqueTerrainsAtVertex returns all unique terrain types that meet at a vertex
// func getUniqueTerrainsAtVertex(tileKey tiles.CoordinateKey, cornerDir mapv1.CornerDirection, tileIndex tiles.Index) []string {
// 	currentTile := tileIndex[tileKey]
// 	terrainSet := make(map[string]bool)
// 	terrainSet[currentTile.TerrainId] = true

// 	// Add neighbor terrains
// 	neighbors := getVertexNeighbors(tileKey, cornerDir)
// 	for _, neighborKey := range neighbors {
// 		if neighbor, exists := tileIndex[neighborKey]; exists {
// 			terrainSet[neighbor.TerrainId] = true
// 		}
// 	}

// 	var terrains []string
// 	for terrain := range terrainSet {
// 		terrains = append(terrains, terrain)
// 	}
// 	return terrains
// }

// // countEdgesAtVertex counts how many edges of the current tile touch the given vertex
// func countEdgesAtVertex(tile *mapv1.Tile, cornerDir mapv1.CornerDirection) int {
// 	// Get the two edge directions that meet at this vertex
// 	edgeDirections := getVertexEdgeDirections(cornerDir)

// 	count := 0
// 	for _, edge := range tile.RenderingSpec.Edges {
// 		for _, edgeDir := range edgeDirections {
// 			if edge.Direction == edgeDir {
// 				count++
// 				break
// 			}
// 		}
// 	}
// 	return count
// }

// // getVertexEdgeDirections returns the two edge directions that meet at a given vertex
// func getVertexEdgeDirections(cornerDir mapv1.CornerDirection) []mapv1.EdgeDirection {
// 	switch cornerDir {
// 	case mapv1.CornerDirection_CORNER_DIRECTION_NW:
// 		return []mapv1.EdgeDirection{mapv1.EdgeDirection_EDGE_DIRECTION_W, mapv1.EdgeDirection_EDGE_DIRECTION_NW}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_N:
// 		return []mapv1.EdgeDirection{mapv1.EdgeDirection_EDGE_DIRECTION_NW, mapv1.EdgeDirection_EDGE_DIRECTION_NE}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_NE:
// 		return []mapv1.EdgeDirection{mapv1.EdgeDirection_EDGE_DIRECTION_NE, mapv1.EdgeDirection_EDGE_DIRECTION_E}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_SE:
// 		return []mapv1.EdgeDirection{mapv1.EdgeDirection_EDGE_DIRECTION_E, mapv1.EdgeDirection_EDGE_DIRECTION_SE}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_S:
// 		return []mapv1.EdgeDirection{mapv1.EdgeDirection_EDGE_DIRECTION_SE, mapv1.EdgeDirection_EDGE_DIRECTION_SW}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_SW:
// 		return []mapv1.EdgeDirection{mapv1.EdgeDirection_EDGE_DIRECTION_SW, mapv1.EdgeDirection_EDGE_DIRECTION_W}
// 	}
// 	return []mapv1.EdgeDirection{}
// }

// // getVertexNeighbors returns the neighbor coordinates that share the given vertex
// func getVertexNeighbors(coord tiles.CoordinateKey, cornerDir mapv1.CornerDirection) []tiles.CoordinateKey {
// 	row := coord.Row
// 	col := coord.Column
// 	depth := coord.Depth
// 	isEvenRow := row%2 == 0

// 	switch cornerDir {
// 	case mapv1.CornerDirection_CORNER_DIRECTION_NW:
// 		if isEvenRow {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row, Column: col - 1},     // W neighbor
// 				{Depth: depth, Row: row - 1, Column: col - 1}, // NW neighbor
// 			}
// 		} else {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row, Column: col - 1}, // W neighbor
// 				{Depth: depth, Row: row - 1, Column: col}, // NW neighbor
// 			}
// 		}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_N:
// 		if isEvenRow {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row - 1, Column: col - 1}, // NW neighbor
// 				{Depth: depth, Row: row - 1, Column: col},     // NE neighbor
// 			}
// 		} else {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row - 1, Column: col},     // NW neighbor
// 				{Depth: depth, Row: row - 1, Column: col + 1}, // NE neighbor
// 			}
// 		}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_NE:
// 		if isEvenRow {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row - 1, Column: col}, // NE neighbor
// 				{Depth: depth, Row: row, Column: col + 1}, // E neighbor
// 			}
// 		} else {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row - 1, Column: col + 1}, // NE neighbor
// 				{Depth: depth, Row: row, Column: col + 1},     // E neighbor
// 			}
// 		}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_SE:
// 		if isEvenRow {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row, Column: col + 1}, // E neighbor
// 				{Depth: depth, Row: row + 1, Column: col}, // SE neighbor
// 			}
// 		} else {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row, Column: col + 1},     // E neighbor
// 				{Depth: depth, Row: row + 1, Column: col + 1}, // SE neighbor
// 			}
// 		}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_S:
// 		if isEvenRow {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row + 1, Column: col},     // SE neighbor
// 				{Depth: depth, Row: row + 1, Column: col - 1}, // SW neighbor
// 			}
// 		} else {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row + 1, Column: col + 1}, // SE neighbor
// 				{Depth: depth, Row: row + 1, Column: col},     // SW neighbor
// 			}
// 		}
// 	case mapv1.CornerDirection_CORNER_DIRECTION_SW:
// 		if isEvenRow {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row + 1, Column: col - 1}, // SW neighbor
// 				{Depth: depth, Row: row, Column: col - 1},     // W neighbor
// 			}
// 		} else {
// 			return []tiles.CoordinateKey{
// 				{Depth: depth, Row: row + 1, Column: col}, // SW neighbor
// 				{Depth: depth, Row: row, Column: col - 1}, // W neighbor
// 			}
// 		}
// 	}
// 	return []tiles.CoordinateKey{}
// }
