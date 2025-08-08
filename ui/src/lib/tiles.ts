import { create } from "@bufbuild/protobuf"
import { type Terrain_RenderingSpec, Terrain_RenderingSpecSchema } from "proto/ts/map/v1/terrain_pb"
import {
    type Segment_Bounds,
    Segment_BoundsSchema,
    type Tile,
    Tile_RenderingSpecSchema,
} from "proto/ts/map/v1/tile_pb"

import { cn } from "./utils"

const emptyBounds = create(Segment_BoundsSchema)

export const boundsIntersect = (
    a: Segment_Bounds = emptyBounds,
    b: Segment_Bounds = emptyBounds,
): boolean => {
    if (a.maxRow < b.minRow || a.minRow > b.maxRow) {
        return false
    }
    if (a.maxColumn < b.minColumn || a.minRow > b.maxRow) {
        return false
    }
    return true
}

export const annotate = (proto: Tile, tileHeight: number, tileWidth: number) => {
    const { row, column } = getCoordinates(proto)
    const even = row % 2 === 0
    const top = row * tileHeight * 0.75
    const left = column * tileWidth + (even ? 0 : tileWidth / 2)

    const terrain = getTerrainRenderingSpec(proto)

    const className = cn(
        "tile",
        "select-none flex items-center justify-center absolute",
        "text-xs",
        terrain.className,
    )

    if (proto.renderingSpec === undefined) {
        proto.renderingSpec = create(Tile_RenderingSpecSchema)
    }

    proto.renderingSpec.top = top
    proto.renderingSpec.left = left
    proto.renderingSpec.className = className
}

export const getCoordinates = (p: Tile) => {
    const row = p.coordinate?.row ?? 0
    const column = p.coordinate?.column ?? 0
    const depth = p.coordinate?.depth ?? 0
    return { row, column, depth }
}

export const getKey = (p: Tile) => {
    const { row, column, depth } = getCoordinates(p)
    return `${row},${column},${depth}`
}

export const boundsInclude = (
    tile: Tile,
    policy: Segment_Bounds = emptyBounds,
    extendBoundsBy = 0,
): boolean => {
    const { row, column } = getCoordinates(tile)

    return (
        row >= policy.minRow - extendBoundsBy &&
        row < policy.maxRow + extendBoundsBy &&
        column >= policy.minColumn - extendBoundsBy &&
        column < policy.maxColumn + extendBoundsBy
    )
}

const ash: Terrain_RenderingSpec = create(Terrain_RenderingSpecSchema, {
    className: "bg-gray-800 hover:bg-gray-900",
})
const grass: Terrain_RenderingSpec = create(Terrain_RenderingSpecSchema, {
    className: "bg-green-800 hover:bg-green-900",
})

export const getTerrainRenderingSpec = (tile: Tile): Terrain_RenderingSpec => {
    switch (tile.terrainId) {
        case "core/terrain/grass":
            return grass
    }
    return ash
}
