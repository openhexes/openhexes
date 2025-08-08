package game

import (
	"context"
	"fmt"
	"slices"
	"time"

	"connectrpc.com/connect"
	"github.com/openhexes/openhexes/api/src/auth"
	"github.com/openhexes/openhexes/api/src/config"
	"github.com/openhexes/openhexes/api/src/server/progress"
	gamev1 "github.com/openhexes/proto/game/v1"
	"github.com/openhexes/proto/game/v1/gamev1connect"
	mapv1 "github.com/openhexes/proto/map/v1"
	progressv1 "github.com/openhexes/proto/progress/v1"
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

func (svc *Service) GetSampleGrid(ctx context.Context, request *connect.Request[gamev1.GetSampleGridRequest], stream *connect.ServerStream[gamev1.GetSampleGridResponse]) error {
	const (
		defaultTotalRows            = uint32(64)
		defaultTotalColumns         = uint32(64)
		defaultMaxRowsPerSegment    = uint32(15)
		defaultMaxColumnsPerSegment = uint32(15)
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

	stageSegments := &progressv1.Stage{
		State: progressv1.Stage_STATE_RUNNING,
		Title: "Prepare segment containers",
	}
	stageGrid := &progressv1.Stage{
		State: progressv1.Stage_STATE_WAITING,
		Title: "Arrange grid",
	}
	stageTiles := &progressv1.Stage{
		State: progressv1.Stage_STATE_WAITING,
		Title: "Process tiles",
	}
	reporter := progress.NewReporter(
		ctx,
		func(p *progressv1.Progress) error {
			return stream.Send(&gamev1.GetSampleGridResponse{
				Progress: p,
			})
		},
		stageSegments, stageGrid, stageTiles,
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

	stageSegments.Duration = durationpb.New(time.Since(start))
	stageSegments.State = progressv1.Stage_STATE_DONE
	stageGrid.State = progressv1.Stage_STATE_RUNNING
	reporter.Update()

	// arrange segments in a grid
	start = time.Now()
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

	for row := range request.Msg.TotalRows {
		segRowIdx := row / request.Msg.MaxRowsPerSegment
		segRow := segmentRows[segRowIdx]

		for column := range request.Msg.TotalColumns {
			segColIdx := column / request.Msg.MaxColumnsPerSegment
			segment := segRow.Segments[segColIdx]

			tile := &mapv1.Tile{
				Coordinate: &mapv1.Tile_Coordinate{
					Row:    uint32(row),
					Column: uint32(column),
				},
			}
			segment.Tiles = append(segment.Tiles, tile)

			processedTileCount++
			if processedTileCount%10_000 == 0 {
				stageTiles.Subtitle = fmt.Sprintf("%d / %d", processedTileCount, totalTiles)
				reporter.Update(float64(processedTileCount) / float64(totalTiles))
			}
		}
	}

	stageTiles.Subtitle = fmt.Sprintf("%d", totalTiles)
	stageTiles.Duration = durationpb.New(time.Since(start))
	stageTiles.State = progressv1.Stage_STATE_DONE
	reporter.Update(1)

	response := &gamev1.GetSampleGridResponse{
		Grid: &mapv1.Grid{
			TotalRows:    request.Msg.TotalRows,
			TotalColumns: request.Msg.TotalColumns,
		},
	}
	if err := stream.Send(response); err != nil {
		return err
	}

	// actually send the grid
	const segmentRowsPerChunk = 10 // todo: smarter way to pick this value
	for rows := range slices.Chunk(segmentRows, segmentRowsPerChunk) {
		response := &gamev1.GetSampleGridResponse{
			Grid: &mapv1.Grid{
				SegmentRows: rows,
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
