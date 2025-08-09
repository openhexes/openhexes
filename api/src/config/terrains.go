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
		"": {
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
		"ash": {
			CanPassWith: defaultCanPassWith,
			CanStopWith: defaultCanStopWith,
			RenderingSpec: &mapv1.Terrain_RenderingSpec{
				RenderingType: mapv1.Terrain_RENDERING_TYPE_ASH,
			},
		},
	}
)
