package config

import mapv1 "github.com/openhexes/proto/map/v1"

var (
	defaultCanPassWith = []mapv1.Terrain_MovementType{
		mapv1.Terrain_MOVEMENT_TYPE_WALKING,
		mapv1.Terrain_MOVEMENT_TYPE_FLYING,
		mapv1.Terrain_MOVEMENT_TYPE_PORTALING,
	}
	defaultCanStopWith = []mapv1.Terrain_MovementType{
		mapv1.Terrain_MOVEMENT_TYPE_WALKING,
		mapv1.Terrain_MOVEMENT_TYPE_FLYING,
		mapv1.Terrain_MOVEMENT_TYPE_PORTALING,
	}
)

var (
	TerrainRegistry = map[string]*mapv1.Terrain{
		"abyss": {
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_ABYSS,
			},
		},
		"water": {
			CanPassWith: []mapv1.Terrain_MovementType{
				mapv1.Terrain_MOVEMENT_TYPE_SWIMMING,
				mapv1.Terrain_MOVEMENT_TYPE_FLYING,
				mapv1.Terrain_MOVEMENT_TYPE_PORTALING,
			},
			CanStopWith: []mapv1.Terrain_MovementType{
				mapv1.Terrain_MOVEMENT_TYPE_SWIMMING,
			},
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_WATER,
			},
		},
		"grass": {
			CanPassWith: defaultCanPassWith,
			CanStopWith: defaultCanStopWith,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_GRASS,
			},
		},
		"highlands": {
			CanPassWith: defaultCanPassWith,
			CanStopWith: defaultCanStopWith,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_HIGHLANDS,
			},
		},
		"dirt": {
			CanPassWith: defaultCanPassWith,
			CanStopWith: defaultCanStopWith,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_DIRT,
			},
		},
		"ash": {
			CanPassWith: defaultCanPassWith,
			CanStopWith: defaultCanStopWith,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_ASH,
			},
		},
		"subterranean": {
			CanPassWith: defaultCanPassWith,
			CanStopWith: defaultCanStopWith,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_SUBTERRANEAN,
			},
		},
		"rough": {
			CanPassWith:     defaultCanPassWith,
			CanStopWith:     defaultCanStopWith,
			MovementPenalty: 25,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_ROUGH,
			},
		},
		"wasteland": {
			CanPassWith:     defaultCanPassWith,
			CanStopWith:     defaultCanStopWith,
			MovementPenalty: 25,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_WASTELAND,
			},
		},
		"sand": {
			CanPassWith:     defaultCanPassWith,
			CanStopWith:     defaultCanStopWith,
			MovementPenalty: 50,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_SAND,
			},
		},
		"snow": {
			CanPassWith:     defaultCanPassWith,
			CanStopWith:     defaultCanStopWith,
			MovementPenalty: 75,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_SNOW,
			},
		},
		"swamp": {
			CanPassWith:     defaultCanPassWith,
			CanStopWith:     defaultCanStopWith,
			MovementPenalty: 75,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_SWAMP,
			},
		},
	}
)

func ValidateRegistries() error {
	for id, v := range TerrainRegistry {
		v.Id = id
	}
	return nil
}
