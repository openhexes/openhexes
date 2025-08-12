package render

import (
	"testing"

	mapv1 "github.com/openhexes/proto/map/v1"
	"github.com/openhexes/openhexes/api/src/tiles"
)

func TestRasterizeSVGToWebP(t *testing.T) {
	// Simple test SVG content
	testSVG := `<svg viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
		<rect x="10" y="10" width="80" height="80" fill="#4f7253"/>
		<polygon points="50,20 80,60 20,60" fill="#3f6543"/>
	</svg>`

	// Test basic rasterization
	webpData, err := rasterizeSVGToWebP(testSVG, 1.0, 80)
	if err != nil {
		t.Fatalf("rasterizeSVGToWebP failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("WebP data is empty")
	}

	// Check WebP magic bytes (first 4 bytes should be "RIFF")
	if len(webpData) < 4 || string(webpData[0:4]) != "RIFF" {
		t.Fatal("Generated data does not appear to be valid WebP (missing RIFF header)")
	}

	// Check for WebP identifier (bytes 8-11 should be "WEBP")
	if len(webpData) < 12 || string(webpData[8:12]) != "WEBP" {
		t.Fatal("Generated data does not appear to be valid WebP (missing WEBP identifier)")
	}

	t.Logf("Successfully generated WebP data: %d bytes", len(webpData))
}

func TestRasterizeSVGToWebPWithScaling(t *testing.T) {
	testSVG := `<svg viewBox="0 0 50 50" xmlns="http://www.w3.org/2000/svg">
		<circle cx="25" cy="25" r="20" fill="#2e4255"/>
	</svg>`

	// Test different scale factors
	scales := []float64{0.5, 1.0, 2.0}
	for _, scale := range scales {
		webpData, err := rasterizeSVGToWebP(testSVG, scale, 70)
		if err != nil {
			t.Fatalf("rasterizeSVGToWebP with scale %f failed: %v", scale, err)
		}

		if len(webpData) == 0 {
			t.Fatalf("WebP data is empty for scale %f", scale)
		}

		t.Logf("Scale %f: generated %d bytes", scale, len(webpData))
	}
}

func TestGenerateWebPSegmentDebug(t *testing.T) {
	// Test the CSS variable replacement
	testSVG := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100">
		<rect x="10" y="10" width="80" height="80" fill="var(--terrain-grass-fill, #78716c)"/>
		<circle cx="50" cy="50" r="20" fill="var(--terrain-water-fill, #2e4255)"/>
	</svg>`
	
	// Test if it can now be rasterized
	_, err := rasterizeSVGToWebP(testSVG, 1.0, 80)
	if err != nil {
		t.Logf("CSS variables SVG failed: %v", err)
	} else {
		t.Logf("CSS variables SVG succeeded")
	}

	// Now test the real generated segment
	segment := &mapv1.Segment{
		Bounds: &mapv1.Segment_Bounds{
			Depth:     0,
			MinRow:    0,
			MaxRow:    1,
			MinColumn: 0,
			MaxColumn: 1,
		},
		Tiles: []*mapv1.Tile{
			{
				Key: "000.000.000",
				Coordinate: &mapv1.Tile_Coordinate{
					Depth:  0,
					Row:    0,
					Column: 0,
				},
				TerrainId: "grass",
				RenderingSpec: &mapv1.Tile_RenderingSpec{
					Edges:   make(map[int32]*mapv1.Tile_Edge),
					Corners: make(map[int32]*mapv1.Tile_Corner),
				},
			},
		},
	}

	tileIndex := make(tiles.Index)
	for _, tile := range segment.Tiles {
		key := tiles.CoordinateToKey(tile.Coordinate)
		tileIndex[key] = tile
	}

	// Test WebP generation with actual segment
	_, err = GenerateWebPSegment(segment, tileIndex, 0)
	if err != nil {
		t.Logf("Real segment WebP generation failed: %v", err)
	} else {
		t.Logf("Real segment WebP generation succeeded!")
	}
}

func TestGenerateWebPSegment(t *testing.T) {
	// Create a test segment with some tiles
	segment := &mapv1.Segment{
		Bounds: &mapv1.Segment_Bounds{
			Depth:     0,
			MinRow:    0,
			MaxRow:    2,
			MinColumn: 0,
			MaxColumn: 2,
		},
		Tiles: []*mapv1.Tile{
			{
				Key: "000.000.000",
				Coordinate: &mapv1.Tile_Coordinate{
					Depth:  0,
					Row:    0,
					Column: 0,
				},
				TerrainId: "grass",
				RenderingSpec: &mapv1.Tile_RenderingSpec{
					Edges:   make(map[int32]*mapv1.Tile_Edge),
					Corners: make(map[int32]*mapv1.Tile_Corner),
				},
			},
			{
				Key: "000.000.001",
				Coordinate: &mapv1.Tile_Coordinate{
					Depth:  0,
					Row:    0,
					Column: 1,
				},
				TerrainId: "water",
				RenderingSpec: &mapv1.Tile_RenderingSpec{
					Edges:   make(map[int32]*mapv1.Tile_Edge),
					Corners: make(map[int32]*mapv1.Tile_Corner),
				},
			},
		},
	}

	// Create a simple tile index
	tileIndex := make(tiles.Index)
	for _, tile := range segment.Tiles {
		key := tiles.CoordinateToKey(tile.Coordinate)
		tileIndex[key] = tile
	}

	// Test WebP generation
	webpData, err := GenerateWebPSegment(segment, tileIndex, 0)
	if err != nil {
		t.Fatalf("GenerateWebPSegment failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Generated WebP data is empty")
	}

	// Verify it's valid WebP
	if len(webpData) < 12 || string(webpData[0:4]) != "RIFF" || string(webpData[8:12]) != "WEBP" {
		t.Fatal("Generated data is not valid WebP format")
	}

	t.Logf("Successfully generated WebP segment: %d bytes", len(webpData))
}

func TestGenerateLightweightWebPSegment(t *testing.T) {
	// Create a test segment
	segment := &mapv1.Segment{
		Bounds: &mapv1.Segment_Bounds{
			Depth:     0,
			MinRow:    0,
			MaxRow:    1,
			MinColumn: 0,
			MaxColumn: 1,
		},
		Tiles: []*mapv1.Tile{
			{
				Key: "000.000.000",
				Coordinate: &mapv1.Tile_Coordinate{
					Depth:  0,
					Row:    0,
					Column: 0,
				},
				TerrainId: "dirt",
				RenderingSpec: &mapv1.Tile_RenderingSpec{
					Edges:   make(map[int32]*mapv1.Tile_Edge),
					Corners: make(map[int32]*mapv1.Tile_Corner),
				},
			},
		},
	}

	tileIndex := make(tiles.Index)
	for _, tile := range segment.Tiles {
		key := tiles.CoordinateToKey(tile.Coordinate)
		tileIndex[key] = tile
	}

	// Test lightweight WebP generation
	webpData, err := GenerateLightweightWebPSegment(segment, tileIndex, 0)
	if err != nil {
		t.Fatalf("GenerateLightweightWebPSegment failed: %v", err)
	}

	if len(webpData) == 0 {
		t.Fatal("Generated lightweight WebP data is empty")
	}

	// Verify it's valid WebP
	if len(webpData) < 12 || string(webpData[0:4]) != "RIFF" || string(webpData[8:12]) != "WEBP" {
		t.Fatal("Generated lightweight data is not valid WebP format")
	}

	t.Logf("Successfully generated lightweight WebP segment: %d bytes", len(webpData))
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"123.45", 123.45},
		{"0", 0.0},
		{"invalid", 0.0}, // Should return 0 on parse error
		{"", 0.0},        // Should return 0 on empty string
	}

	for _, test := range tests {
		result := parseFloat(test.input)
		if result != test.expected {
			t.Errorf("parseFloat(%q) = %f, expected %f", test.input, result, test.expected)
		}
	}
}