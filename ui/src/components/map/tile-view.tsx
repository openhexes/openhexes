import { useTileDimensions } from "@/hooks/use-tiles"
import { getCoordinates, getTerrainRenderingSpec } from "@/lib/tiles"
import { cn } from "@/lib/utils"
import type { Tile } from "proto/ts/map/v1/tile_pb"
import React from "react"

import "./tile.css"

export interface TileProps {
    tile: Tile
}

export const TileView: React.FC<TileProps> = ({ tile }) => {
    const { tileHeight, tileWidth } = useTileDimensions()

    const { row, column } = getCoordinates(tile)
    const even = row % 2 === 0
    const top = row * tileHeight * 0.75
    const left = column * tileWidth + (even ? 0 : tileWidth / 2)

    const terrain = getTerrainRenderingSpec(tile)

    const className = cn(
        "tile",
        "select-none flex items-center justify-center absolute",
        "text-xs",
        terrain.className,
        "bg-transparent",
    )

    const style: React.CSSProperties = {
        top,
        left,
        height: tileHeight,
        fontSize: 8,
        "--tile-x": `${left}px`,
        "--tile-y": `${top}px`,
    } as React.CSSProperties

    return <div className={cn(className)} style={style} />
}

export default TileView
