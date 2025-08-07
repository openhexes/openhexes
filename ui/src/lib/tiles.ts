import { create } from "@bufbuild/protobuf"
import { type Terrain_RenderingSpec, Terrain_RenderingSpecSchema } from "proto/ts/map/v1/terrain_pb"
import type { Tile } from "proto/ts/map/v1/tile_pb"

import { cn } from "./utils"

export interface Bounds {
    minRow: number
    maxRow: number
    minColumn: number
    maxColumn: number
}

export const getBoundsKey = (b: Bounds): string => {
    return `ROW[${b.minRow},${b.maxRow}) COL[${b.minColumn},${b.maxColumn})`
}

export const boundsIntersect = (a: Bounds, b: Bounds): boolean => {
    if (a.maxRow < b.minRow || a.minRow > b.maxRow) {
        return false
    }
    if (a.maxColumn < b.minColumn || a.minRow > b.maxRow) {
        return false
    }
    return true
}

export interface AnnotatedTile {
    proto: Tile
    height: number
    top: number
    left: number
    className: string
}

export interface Segment {
    tiles: AnnotatedTile[]
    bounds: Bounds
}

export interface Grid {
    segmentRows: Segment[][]
    totalRows: number
    totalColumns: number
}

export const annotate = (proto: Tile, tileHeight: number, tileWidth: number): AnnotatedTile => {
    const { row, column } = getCoordinates(proto)
    const even = row % 2 === 0
    const top = row * tileHeight * 0.75
    const left = column * tileWidth + (even ? 0 : tileWidth / 2)

    const spec = getRenderingSpec(proto)

    const className = cn(
        "tile",
        "select-none flex items-center justify-center absolute",
        "text-xs",
        spec.color,
    )

    return {
        proto,
        height: tileHeight,
        top,
        left,
        className,
    }
}

export const emptyBounds: Bounds = {
    minRow: 0,
    maxRow: 0,
    minColumn: 0,
    maxColumn: 0,
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
    policy: Bounds = emptyBounds,
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
    color: "bg-gray-800 hover:bg-gray-900",
})
const grass: Terrain_RenderingSpec = create(Terrain_RenderingSpecSchema, {
    color: "bg-green-800 hover:bg-green-900",
})

export const getRenderingSpec = (tile: Tile): Terrain_RenderingSpec => {
    switch (tile.terrainId) {
        case "core/terrain/grass":
            return grass
    }
    return ash
}
