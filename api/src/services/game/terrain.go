package game

import (
	"math"
	"math/rand"
	"time"

	"github.com/openhexes/openhexes/api/src/tiles"
)

// Simple noise implementation for terrain generation
func simpleNoise(x, y float64) float64 {
	return math.Sin(x*0.1)*math.Cos(y*0.1) +
		0.5*math.Sin(x*0.2)*math.Cos(y*0.2) +
		0.25*math.Sin(x*0.4)*math.Cos(y*0.4)
}

// Generate realistic terrain based on heightmap and moisture
func generateRealisticTerrain(totalRows, totalColumns, depth, totalLayers uint32) map[tiles.CoordinateKey]string {
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

			// Determine terrain type based on height, moisture, temperature, and depth
			ck := tiles.CoordinateKey{Depth: depth, Row: row, Column: column}

			// Calculate depth progression as a percentage of total layers (0.0 to 1.0)
			depthProgress := float64(depth) / math.Max(float64(totalLayers-1), 1.0)
			
			// More gradual depth influence - scales with total layers
			depthFactor := depthProgress * 0.4 // Reduced from 0.3 * depth to gradual progression
			adjustedHeight := height - depthFactor

			switch {
			case adjustedHeight < -0.7:
				// Abyss appears only in the deepest areas of deeper layers
				abyssChance := depthProgress * 0.5 // Max 50% chance in deepest layer
				if depthProgress > 0.6 && rng.Float64() < abyssChance {
					terrainMap[ck] = "abyss"
				} else if depthProgress > 0.3 && rng.Float64() < depthProgress*0.8 {
					terrainMap[ck] = "subterranean"
				} else {
					terrainMap[ck] = "water"
				}
			case adjustedHeight < -0.1:
				// Surface layers prefer water, deeper layers prefer subterranean and dirt
				subterraneanChance := depthProgress * 0.6
				dirtChance := depthProgress * 0.4
				if rng.Float64() < subterraneanChance {
					terrainMap[ck] = "subterranean"
				} else if rng.Float64() < dirtChance {
					terrainMap[ck] = "dirt"
				} else {
					terrainMap[ck] = "water"
				}
			case adjustedHeight < 0.2:
				subterraneanChance := depthProgress * 0.5
				dirtChance := depthProgress * 0.7
				if rng.Float64() < subterraneanChance {
					terrainMap[ck] = "subterranean"
				} else if rng.Float64() < dirtChance {
					terrainMap[ck] = "dirt"
				} else if depthProgress < 0.3 {
					// Surface and shallow layers: preserve original logic
					if moisture > 0.3 {
						terrainMap[ck] = "swamp"
					} else {
						terrainMap[ck] = "water"
					}
				} else {
					// Mid-depth layers: mix of terrain
					if moisture > 0.3 && rng.Float64() < 0.3 {
						terrainMap[ck] = "swamp"
					} else if rng.Float64() < 0.4 {
						terrainMap[ck] = "dirt"
					} else {
						terrainMap[ck] = "water"
					}
				}
			case adjustedHeight < 0.6:
				subterraneanChance := depthProgress * 0.6
				dirtChance := depthProgress * 0.5
				if rng.Float64() < subterraneanChance {
					terrainMap[ck] = "subterranean"
				} else if rng.Float64() < dirtChance {
					terrainMap[ck] = "dirt"
				} else if depthProgress < 0.2 {
					// Surface layer: preserve original logic favoring grass and water
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
				} else {
					// Deeper layers: balanced mix including some surface terrain
					if rng.Float64() < 0.3 {
						terrainMap[ck] = "dirt"
					} else if moisture < -0.1 && rng.Float64() < 0.2 {
						terrainMap[ck] = "subterranean"
					} else if temperature < 0.2 {
						terrainMap[ck] = "snow"
					} else {
						terrainMap[ck] = "grass"
					}
				}
			case adjustedHeight < 1.0:
				subterraneanChance := depthProgress * 0.7
				if rng.Float64() < subterraneanChance {
					terrainMap[ck] = "subterranean"
				} else if depthProgress > 0.5 && rng.Float64() < 0.3 {
					terrainMap[ck] = "dirt"
				} else if depthProgress < 0.2 {
					// Surface layer: preserve original highland logic
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
				} else {
					// Mid to deep layers: mix of terrain
					if rng.Float64() < 0.3 {
						terrainMap[ck] = "dirt"
					} else if temperature < 0.3 {
						terrainMap[ck] = "snow"
					} else {
						terrainMap[ck] = "highlands"
					}
				}
			case adjustedHeight < 1.4:
				subterraneanChance := depthProgress * 0.8
				if rng.Float64() < subterraneanChance {
					terrainMap[ck] = "subterranean"
				} else if depthProgress < 0.2 {
					// Surface layer: preserve original logic
					if temperature < 0.4 {
						terrainMap[ck] = "snow"
					} else if moisture < -0.1 {
						terrainMap[ck] = "wasteland"
					} else if rng.Float64() < 0.3 {
						terrainMap[ck] = "ash"
					} else {
						terrainMap[ck] = "highlands"
					}
				} else {
					// Deeper layers: varied terrain with subterranean preference
					if rng.Float64() < 0.4 {
						terrainMap[ck] = "highlands"
					} else if rng.Float64() < 0.3 {
						terrainMap[ck] = "ash"
					} else {
						terrainMap[ck] = "dirt"
					}
				}
			default:
				// Very high terrain: abyss only in deepest layers, subterranean common
				abyssChance := math.Max(0, (depthProgress-0.7)*1.5) // Only in last 30% of layers, scaling up
				if rng.Float64() < abyssChance {
					terrainMap[ck] = "abyss"
				} else if depthProgress > 0.3 {
					terrainMap[ck] = "subterranean"
				} else {
					// Surface layer: preserve original logic
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
	}

	return terrainMap
}