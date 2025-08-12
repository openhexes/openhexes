package render

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	"github.com/openhexes/openhexes/api/src/tiles"
	mapv1 "github.com/openhexes/proto/map/v1"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// rasterizeSVGToWebP converts SVG string to WebP bytes at specified scale
func rasterizeSVGToWebP(svgContent string, scaleFactor float64, quality int) ([]byte, error) {
	if svgContent == "" {
		return nil, fmt.Errorf("empty SVG content")
	}

	// Parse SVG
	icon, err := oksvg.ReadIconStream(strings.NewReader(svgContent), oksvg.StrictErrorMode)
	if err != nil {
		return nil, fmt.Errorf("parsing SVG: %w", err)
	}

	// Use SVG viewBox dimensions if available, otherwise use default calculations
	targetWidth := int(800.0 * scaleFactor)  // Default segment width
	targetHeight := int(600.0 * scaleFactor) // Default segment height

	// Try to extract viewBox from SVG for more accurate dimensions
	if viewBoxStart := strings.Index(svgContent, "viewBox=\""); viewBoxStart != -1 {
		viewBoxStart += len("viewBox=\"")
		if viewBoxEnd := strings.Index(svgContent[viewBoxStart:], "\""); viewBoxEnd != -1 {
			viewBox := svgContent[viewBoxStart : viewBoxStart+viewBoxEnd]
			parts := strings.Fields(viewBox)
			if len(parts) >= 4 {
				// viewBox format: "x y width height"
				if w := parseFloat(parts[2]); w > 0 {
					targetWidth = int(w * scaleFactor)
				}
				if h := parseFloat(parts[3]); h > 0 {
					targetHeight = int(h * scaleFactor)
				}
			}
		}
	}

	// Set SVG target size
	icon.SetTarget(0, 0, float64(targetWidth), float64(targetHeight))

	// Create RGBA image
	rgba := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))

	// Create scanner and rasterize SVG
	scanner := rasterx.NewScannerGV(targetWidth, targetHeight, rgba, rgba.Bounds())
	dasher := rasterx.NewDasher(targetWidth, targetHeight, scanner)
	icon.Draw(dasher, 1.0)

	// Encode to WebP
	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, rgba, nil); err != nil {
		return nil, fmt.Errorf("failed to encode WebP: %w", err)
	}

	return buf.Bytes(), nil
}

// parseFloat attempts to parse a float from a string, returns 0 on error
func parseFloat(s string) float64 {
	f := 0.0
	fmt.Sscanf(s, "%f", &f)
	return f
}

// GenerateWebPSegment generates a WebP by rasterizing the existing SVG
func GenerateWebPSegment(segment *mapv1.Segment, tileIndex tiles.Index, layerDepth uint32) ([]byte, error) {
	// Get the existing SVG content - this already has all the correct edges, corners, and patterns
	svgContent, err := GenerateSVGSegment(segment, tileIndex, layerDepth)
	if err != nil {
		return nil, fmt.Errorf("generating segment SVG: %s: %w", segment.Key, err)
	}

	// Rasterize SVG to high-quality WebP (2x scale for crisp rendering)
	return rasterizeSVGToWebP(svgContent, 2.0, 80)
}

// GenerateLightweightWebPSegment generates a lightweight WebP by rasterizing the lightweight SVG
func GenerateLightweightWebPSegment(segment *mapv1.Segment, tileIndex tiles.Index, layerDepth uint32) ([]byte, error) {
	// Get the existing lightweight SVG content
	svgContent, err := GenerateLightweightSVGSegment(segment, tileIndex, layerDepth)
	if err != nil {
		return nil, fmt.Errorf("generating segment SVG (lightweight): %s: %w", segment.Key, err)
	}

	// Rasterize lightweight SVG to WebP (1x scale for performance, higher compression)
	return rasterizeSVGToWebP(svgContent, 1.0, 60)
}
