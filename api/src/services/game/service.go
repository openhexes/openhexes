package game

import (
	"context"
	"fmt"
	"math"
	"math/rand"
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

// Simple noise implementation for terrain generation
func simpleNoise(x, y float64) float64 {
	return math.Sin(x*0.1) * math.Cos(y*0.1) + 
		   0.5*math.Sin(x*0.2) * math.Cos(y*0.2) + 
		   0.25*math.Sin(x*0.4) * math.Cos(y*0.4)
}

// Generate realistic terrain based on heightmap and moisture
func generateRealisticTerrain(totalRows, totalColumns uint32) map[tiles.CoordinateKey]string {
	terrainMap := make(map[tiles.CoordinateKey]string)
	
	// Random seed for each generation - creates different maps each time
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	seedOffset := rng.Float64() * 10000 // Random offset for noise patterns
	
	for row := uint32(0); row < totalRows; row++ {
		for column := uint32(0); column < totalColumns; column++ {
			x, y := float64(column), float64(row)
			
			// Generate heightmap using multiple octaves of noise with random seed
			height := 0.6*simpleNoise((x+seedOffset)*0.008, (y+seedOffset)*0.008) + 
					 0.4*simpleNoise((x+seedOffset)*0.02, (y+seedOffset)*0.02) + 
					 0.3*simpleNoise((x+seedOffset)*0.05, (y+seedOffset)*0.05) + 
					 0.2*simpleNoise((x+seedOffset)*0.1, (y+seedOffset)*0.1)
			
			// Generate moisture map with different seed offset
			moistureSeed := seedOffset + 1000
			moisture := 0.5*simpleNoise((x+moistureSeed)*0.015, (y+moistureSeed)*0.015) + 
					   0.3*simpleNoise((x+moistureSeed)*0.04, (y+moistureSeed)*0.04) +
					   0.2*simpleNoise((x+moistureSeed)*0.08, (y+moistureSeed)*0.08)
			
			// Generate temperature (affected by latitude + noise)
			tempSeed := seedOffset + 2000
			temperature := 0.9 - 0.7*float64(row)/float64(totalRows) + 
						  0.3*simpleNoise((x+tempSeed)*0.025, (y+tempSeed)*0.025) +
						  0.2*simpleNoise((x+tempSeed)*0.06, (y+tempSeed)*0.06)
			
			// Add controlled randomness
			randomFactor := (rng.Float64() - 0.5) * 0.3
			height += randomFactor
			
			// Determine terrain type based on height, moisture, and temperature
			ck := tiles.CoordinateKey{Depth: 0, Row: row, Column: column}
			
			switch {
			case height < -0.7:
				terrainMap[ck] = "abyss"
			case height < -0.1:
				terrainMap[ck] = "water"
			case height < 0.2:
				if moisture > 0.3 {
					terrainMap[ck] = "swamp"
				} else {
					terrainMap[ck] = "water"
				}
			case height < 0.6:
				if temperature < 0.2 {
					terrainMap[ck] = "snow"
				} else if moisture > 0.4 {
					terrainMap[ck] = "swamp"
				} else if moisture < -0.3 && temperature > 0.7 {
					terrainMap[ck] = "sand"
				} else if moisture < -0.1 {
					terrainMap[ck] = "dirt"
				} else {
					terrainMap[ck] = "grass"
				}
			case height < 1.0:
				if temperature < 0.3 {
					terrainMap[ck] = "snow"
				} else if moisture < -0.2 && temperature > 0.6 {
					terrainMap[ck] = "sand"
				} else if moisture < 0.0 {
					if rng.Float64() < 0.3 {
						terrainMap[ck] = "wasteland"
					} else {
						terrainMap[ck] = "dirt"
					}
				} else if rng.Float64() < 0.2 {
					terrainMap[ck] = "rough"
				} else {
					terrainMap[ck] = "highlands"
				}
			case height < 1.4:
				if temperature < 0.4 {
					terrainMap[ck] = "snow"
				} else if moisture < -0.1 {
					terrainMap[ck] = "wasteland"
				} else if rng.Float64() < 0.3 {
					terrainMap[ck] = "ash"
				} else {
					terrainMap[ck] = "highlands"
				}
			default:
				if temperature < 0.5 {
					terrainMap[ck] = "snow"
				} else if rng.Float64() < 0.5 {
					terrainMap[ck] = "ash"
				} else {
					terrainMap[ck] = "subterranean"
				}
			}
		}
	}
	
	return terrainMap
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

	// Generate realistic terrain using noise-based heightmap
	terrainMap := generateRealisticTerrain(request.Msg.TotalRows, request.Msg.TotalColumns)

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
			tile.Key = fmt.Sprintf("%03d.%03d.%03d", k.Depth, k.Row, k.Column)
			idx[k] = tile

			if terrain, ok := terrainMap[k]; ok {
				tile.TerrainId = terrain
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
			Edges:   make(map[int32]*mapv1.Tile_Edge, 6),
			Corners: make(map[int32]*mapv1.Tile_Corner, 6),
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

			tile.RenderingSpec.Edges[int32(c.Direction)] = &mapv1.Tile_Edge{
				Direction:          c.Direction,
				NeighbourTerrainId: neighbour.TerrainId,
			}
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
			cns := tiles.GetCornerNeighbours(k, cd)

			for _, cn := range cns {
				n, ok := idx[cn.CoordinateKey]
				if !ok {
					// that's okay for tiles on the edge of the map
					continue
				}

				opCD, opED := tiles.GetOppositeCorner(cd, cn.EdgeDirection)
				opE, ok := n.RenderingSpec.Edges[int32(opED)]
				if !ok {
					// neighbour doesn't have a connecting edge
					continue
				}

				corner, ok := n.RenderingSpec.Corners[int32(opCD)]
				if !ok {
					corner = &mapv1.Tile_Corner{
						Direction: opCD,
						Edges:     make(map[int32]*mapv1.Tile_Edge, 2),
					}
					n.RenderingSpec.Corners[int32(opCD)] = corner
				}
				corner.Edges[int32(opED)] = opE
			}
		}

		processedTileCount++
		if processedTileCount%10_000 == 0 {
			stageCorners.Subtitle = fmt.Sprintf("%d / %d", processedTileCount, totalTiles)
			reporter.Update(float64(processedTileCount) / float64(totalTiles))
		}
	}

	// remove extra corners between two existing edges of the same terrain
	for k, tile := range idx {
	CornerDirections:
		for _, cd := range tiles.AllCornerDirections {
			cornerNeighbours := tiles.GetCornerNeighbours(k, cd)
			if len(cornerNeighbours) != 2 {
				continue
			}

			existingEdges := make(map[mapv1.EdgeDirection]struct{}, 2)
			existingTerrains := make(map[string]struct{}, 2)
			for _, e := range tile.RenderingSpec.Edges {
				existingEdges[e.Direction] = struct{}{}
				existingTerrains[e.NeighbourTerrainId] = struct{}{}
			}

			if len(existingTerrains) > 1 {
				continue
			}

			for _, n := range cornerNeighbours {
				if _, ok := existingEdges[n.EdgeDirection]; !ok {
					continue CornerDirections
				}
			}

			// both edges are present, corner is not needed
			delete(tile.RenderingSpec.Corners, int32(cd))
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
	if err := stream.Send(terrainsResponse); err != nil {
		return err
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
