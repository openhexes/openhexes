package game

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
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
	return math.Sin(x*0.1)*math.Cos(y*0.1) +
		0.5*math.Sin(x*0.2)*math.Cos(y*0.2) +
		0.25*math.Sin(x*0.4)*math.Cos(y*0.4)
}

// Generate realistic terrain based on heightmap and moisture
func generateRealisticTerrain(totalRows, totalColumns, depth uint32) map[tiles.CoordinateKey]string {
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
			ck := tiles.CoordinateKey{Depth: depth, Row: row, Column: column}

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

func generateLayer(depth uint32, totalRows, totalColumns, maxRowsPerSegment, maxColumnsPerSegment uint32) *mapv1.Layer {
	layer := &mapv1.Layer{
		Depth:        depth,
		TotalRows:    totalRows,
		TotalColumns: totalColumns,
		SegmentRows:  make([]*mapv1.Segment_Row, 0, totalRows/maxRowsPerSegment),
	}

	// Step 1: Generate terrain for this layer
	terrainMap := generateRealisticTerrain(totalRows, totalColumns, depth)

	// Step 2: Create per-layer tile index
	layerTileCount := totalRows * totalColumns
	layerIdx := make(tiles.Index, layerTileCount)

	// Step 3: Create tiles for this layer
	for row := range totalRows {
		for column := range totalColumns {
			k := tiles.CoordinateKey{Depth: depth, Row: uint32(row), Column: uint32(column)}

			tile := &mapv1.Tile{
				Key: fmt.Sprintf("%03d.%03d.%03d", depth, row, column),
				Coordinate: &mapv1.Tile_Coordinate{
					Depth:  depth,
					Row:    uint32(row),
					Column: uint32(column),
				},
			}

			if terrain, ok := terrainMap[k]; ok {
				tile.TerrainId = terrain
			} else {
				tile.TerrainId = "water"
			}

			layerIdx[k] = tile
		}
	}

	// Step 4: Calculate edges for this layer
	for k, tile := range layerIdx {
		tile.RenderingSpec = &mapv1.Tile_RenderingSpec{
			Edges:   make(map[int32]*mapv1.Tile_Edge, 6),
			Corners: make(map[int32]*mapv1.Tile_Corner, 6),
		}
		tileTerrain, ok := config.TerrainRegistry[tile.TerrainId]
		if !ok {
			continue
		}

		for c := range tiles.IterNeighbours(k) {
			neighbour, ok := layerIdx[c.CoordinateKey]
			if !ok || neighbour.TerrainId == tile.TerrainId {
				continue
			}

			neighbourTerrain, ok := config.TerrainRegistry[neighbour.TerrainId]
			if !ok {
				continue
			}

			if tileTerrain.RenderingSpec.RenderingType > neighbourTerrain.RenderingSpec.RenderingType {
				continue
			}

			tile.RenderingSpec.Edges[int32(c.Direction)] = &mapv1.Tile_Edge{
				Direction:          c.Direction,
				NeighbourTerrainId: neighbour.TerrainId,
			}
		}
	}

	// Step 5: Calculate corners for this layer
	for k := range layerIdx {
		for _, cd := range tiles.AllCornerDirections {
			cns := tiles.GetCornerNeighbours(k, cd)

			for _, cn := range cns {
				n, ok := layerIdx[cn.CoordinateKey]
				if !ok {
					continue
				}

				opCD, opED := tiles.GetOppositeCorner(cd, cn.EdgeDirection)
				opE, ok := n.RenderingSpec.Edges[int32(opED)]
				if !ok {
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
	}

	// Step 6: Remove extra corners between two existing edges of the same terrain
	for k, tile := range layerIdx {
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

	// Step 7: Create segments and segment grid for this layer
	segmentsLength := totalRows / maxRowsPerSegment * totalColumns / maxColumnsPerSegment
	segments := make([]*mapv1.Segment, 0, segmentsLength)

	for rowStart := uint32(0); rowStart < totalRows; rowStart += maxRowsPerSegment {
		for columnStart := uint32(0); columnStart < totalColumns; columnStart += maxColumnsPerSegment {
			rowEnd := rowStart + maxRowsPerSegment
			columnEnd := columnStart + maxColumnsPerSegment

			tilesCapacity := maxRowsPerSegment * maxColumnsPerSegment

			segments = append(segments, &mapv1.Segment{
				Tiles: make([]*mapv1.Tile, 0, tilesCapacity),
				Bounds: &mapv1.Segment_Bounds{
					MinRow:    int32(rowStart),
					MaxRow:    int32(rowEnd),
					MinColumn: int32(columnStart),
					MaxColumn: int32(columnEnd),
				},
			})
		}
	}

	// Step 8: Assign tiles to segments for this layer
	for row := range totalRows {
		segRowIdx := row / maxRowsPerSegment
		for column := range totalColumns {
			segColIdx := column / maxColumnsPerSegment
			segmentIndex := segRowIdx*(totalColumns/maxColumnsPerSegment) + segColIdx
			segment := segments[segmentIndex]

			coordinate := &mapv1.Tile_Coordinate{
				Depth:  depth,
				Row:    uint32(row),
				Column: uint32(column),
			}
			k := tiles.CoordinateToKey(coordinate)
			tile := layerIdx[k]

			segment.Tiles = append(segment.Tiles, tile)
		}
	}

	// Step 9: Arrange segments in a grid for this layer
	gridRowLength := totalRows / maxRowsPerSegment
	segmentsPerRow := totalColumns / maxColumnsPerSegment
	gridRow := make([]*mapv1.Segment, 0, segmentsPerRow)
	var rowStart *int32

	for _, segment := range segments {
		if rowStart == nil {
			rowStart = &segment.Bounds.MinRow
			gridRow = make([]*mapv1.Segment, 0, gridRowLength)
			gridRow = append(gridRow, segment)
		} else if *rowStart != segment.Bounds.MinRow {
			layer.SegmentRows = append(layer.SegmentRows, &mapv1.Segment_Row{Segments: gridRow})
			gridRow = make([]*mapv1.Segment, 0, gridRowLength)
			gridRow = append(gridRow, segment)
			rowStart = &segment.Bounds.MinRow
		} else {
			gridRow = append(gridRow, segment)
		}
	}
	layer.SegmentRows = append(layer.SegmentRows, &mapv1.Segment_Row{Segments: gridRow})

	// Step 10: Render SVG for all segments in this layer
	for _, row := range layer.SegmentRows {
		for _, segment := range row.Segments {
			segment.RenderingSpec = &mapv1.Segment_RenderingSpec{
				Svg: render.GenerateSVGSegment(segment, layerIdx, depth),
			}
		}
	}

	return layer
}

func sendWorld(stream *connect.ServerStream[gamev1.GetSampleWorldResponse], request *connect.Request[gamev1.GetSampleWorldRequest], world *worldv1.World) error {
	const defaultMaxChunkSizeBytes = 32 * 1024 // 32Kb

	dimensionsResponse := &gamev1.GetSampleWorldResponse{
		World: &worldv1.World{
			RenderingSpec: &worldv1.World_RenderingSpec{
				TileHeight: config.TileHeight,
				TileWidth:  config.TileWidth,
			},
			Layers: make([]*mapv1.Layer, 0, len(world.Layers)),
		},
	}
	for _, layer := range world.Layers {
		dimensionsResponse.World.Layers = append(dimensionsResponse.World.Layers, &mapv1.Layer{
			Depth:        layer.Depth,
			Name:         layer.Name,
			TotalRows:    layer.TotalRows,
			TotalColumns: layer.TotalColumns,
		})
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

	// tiles - send each layer separately
	for _, layer := range world.Layers {
		chunk := &mapv1.Layer{
			Depth:       layer.Depth,
			SegmentRows: make([]*mapv1.Segment_Row, 0, 100),
		}
		for _, row := range layer.SegmentRows {
			chunk.SegmentRows = append(chunk.SegmentRows, row)
			if proto.Size(chunk) >= defaultMaxChunkSizeBytes {
				response := &gamev1.GetSampleWorldResponse{
					World: &worldv1.World{
						Layers: []*mapv1.Layer{chunk},
					},
				}
				if err := stream.Send(response); err != nil {
					return err
				}
				chunk = &mapv1.Layer{
					Depth:       layer.Depth,
					SegmentRows: make([]*mapv1.Segment_Row, 0, 100),
				}
			}
		}
		if len(chunk.SegmentRows) > 0 {
			response := &gamev1.GetSampleWorldResponse{
				World: &worldv1.World{
					Layers: []*mapv1.Layer{chunk},
				},
			}
			if err := stream.Send(response); err != nil {
				return err
			}
		}
	}

	return nil
}

func (svc *Service) GetSampleWorld(ctx context.Context, request *connect.Request[gamev1.GetSampleWorldRequest], stream *connect.ServerStream[gamev1.GetSampleWorldResponse]) error {
	const (
		defaultTotalRows            = uint32(64)
		defaultTotalColumns         = uint32(64)
		defaultMaxRowsPerSegment    = uint32(15)
		defaultMaxColumnsPerSegment = uint32(15)
		defalutTotalLayers          = uint32(2)
	)

	if request.Msg.TotalLayers == uint32(0) {
		request.Msg.TotalLayers = defalutTotalLayers
	}
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

	// Generate all layers in parallel
	start := time.Now()
	world := &worldv1.World{
		Layers: make([]*mapv1.Layer, request.Msg.TotalLayers),
	}

	// Comment out progress reporting as requested
	// Progress reporting will be made more suitable for parallelized generation later

	var wg sync.WaitGroup

	// Spawn one goroutine per layer, each handling the complete pipeline
	for depth := uint32(0); depth < request.Msg.TotalLayers; depth++ {
		wg.Add(1)
		go func(d uint32) {
			defer wg.Done()
			world.Layers[d] = generateLayer(d, request.Msg.TotalRows, request.Msg.TotalColumns, request.Msg.MaxRowsPerSegment, request.Msg.MaxColumnsPerSegment)
		}(depth)
	}

	// Wait for all layers to complete
	wg.Wait()

	// Update stages after all work is done
	stageGrid.Duration = durationpb.New(time.Since(start))
	stageGrid.State = progressv1.Stage_STATE_DONE
	stageTiles.Duration = durationpb.New(time.Since(start))
	stageTiles.State = progressv1.Stage_STATE_DONE
	stageEdges.Duration = durationpb.New(time.Since(start))
	stageEdges.State = progressv1.Stage_STATE_DONE
	stageCorners.Duration = durationpb.New(time.Since(start))
	stageCorners.State = progressv1.Stage_STATE_DONE
	stageRender.Duration = durationpb.New(time.Since(start))
	stageRender.State = progressv1.Stage_STATE_DONE
	reporter.Update(1)

	// Send the world to the client
	return sendWorld(stream, request, world)
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
