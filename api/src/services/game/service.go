package game

import (
	"context"
	"errors"
	"fmt"
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

func New(cfg *config.Config, auth *auth.Controller) *Service {
	return &Service{
		cfg:  cfg,
		auth: auth,
	}
}

func generateLayer(depth uint32, totalRows, totalColumns, maxRowsPerSegment, maxColumnsPerSegment, totalLayers uint32) (*mapv1.Layer, error) {
	layer := &mapv1.Layer{
		Depth:        depth,
		TotalRows:    totalRows,
		TotalColumns: totalColumns,
		SegmentRows:  make([]*mapv1.Segment_Row, 0, totalRows/maxRowsPerSegment),
	}

	// Step 1: Generate terrain for this layer
	terrainMap := generateRealisticTerrain(totalRows, totalColumns, depth, totalLayers)

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
	segmentRowsCount := (totalRows + maxRowsPerSegment - 1) / maxRowsPerSegment             // ceiling division
	segmentColumnsCount := (totalColumns + maxColumnsPerSegment - 1) / maxColumnsPerSegment // ceiling division
	segmentsLength := segmentRowsCount * segmentColumnsCount
	segments := make([]*mapv1.Segment, 0, segmentsLength)

	for rowStart := uint32(0); rowStart < totalRows; rowStart += maxRowsPerSegment {
		for columnStart := uint32(0); columnStart < totalColumns; columnStart += maxColumnsPerSegment {
			rowEnd := rowStart + maxRowsPerSegment
			columnEnd := columnStart + maxColumnsPerSegment

			tilesCapacity := maxRowsPerSegment * maxColumnsPerSegment

			segments = append(segments, &mapv1.Segment{
				Key:   fmt.Sprintf("%d.%d.%d", depth, rowStart, columnStart),
				Tiles: make([]*mapv1.Tile, 0, tilesCapacity),
				Bounds: &mapv1.Segment_Bounds{
					Depth:     depth,
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
			segmentsPerRow := (totalColumns + maxColumnsPerSegment - 1) / maxColumnsPerSegment // ceiling division
			segmentIndex := segRowIdx*segmentsPerRow + segColIdx
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
			segment.RenderingSpec = &mapv1.Segment_RenderingSpec{}
			var err error
			segment.RenderingSpec.Svg, err = render.GenerateSVGSegment(segment, layerIdx, depth)
			if err != nil {
				return nil, fmt.Errorf("generating SVG for segment: %s: %w", segment.Key, err)
			}
			segment.RenderingSpec.SvgLightweight, err = render.GenerateLightweightSVGSegment(segment, layerIdx, depth)
			if err != nil {
				return nil, fmt.Errorf("generating SVG (lightweight) for segment: %s: %w", segment.Key, err)
			}
		}
	}

	return layer, nil
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
	errChan := make(chan error, request.Msg.TotalLayers)

	// Spawn one goroutine per layer, each handling the complete pipeline
	for depth := uint32(0); depth < request.Msg.TotalLayers; depth++ {
		wg.Add(1)
		go func(d uint32) {
			defer wg.Done()
			layer, err := generateLayer(d, request.Msg.TotalRows, request.Msg.TotalColumns, request.Msg.MaxRowsPerSegment, request.Msg.MaxColumnsPerSegment, request.Msg.TotalLayers)
			if err != nil {
				errChan <- fmt.Errorf("[depth=%d] %w", d, err)
				return
			}
			world.Layers[d] = layer
		}(depth)
	}

	// Wait for all layers to complete
	wg.Wait()
	close(errChan)

	// Check for any errors
	var err error
	for e := range errChan {
		err = errors.Join(err, e)
	}
	if err != nil {
		return fmt.Errorf("generating layers: %w", err)
	}

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
