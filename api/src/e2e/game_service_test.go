package e2e

import (
	"testing"

	"github.com/openhexes/openhexes/api/src/test"
	gamev1 "github.com/openhexes/proto/game/v1"
)

func TestGameService(t *testing.T) {
	e := test.NewEnv(t)

	e.Run("get custom sample world", func(e *test.Env) {
		token := e.Config.Test.Tokens.Owner
		request := &gamev1.GetSampleWorldRequest{
			TotalLayers:          3,
			TotalRows:            8,
			TotalColumns:         8,
			MaxRowsPerSegment:    4,
			MaxColumnsPerSegment: 4,
		}

		world := e.GetSampleWorld(request, test.WithToken(token))
		e.Require.Len(world.Layers, int(request.TotalLayers))

		// Verify each layer has the expected structure
		for i, layer := range world.Layers {
			e.Require.NotNil(layer, "layer %d should not be nil", i)
			e.Require.Equal(request.TotalRows, layer.TotalRows, "layer %d should have 8 rows", i)
			e.Require.Equal(request.TotalColumns, layer.TotalColumns, "layer %d should have 8 columns", i)
		}
	})
}
