import { useTileDimensions } from "@/hooks/use-tiles"
import { getCoordinates } from "@/lib/tiles"
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

    const className = cn(
        "tile",
        "select-none flex items-center justify-center absolute",
        "text-xs text-transparent bg-transparent hover:bg-gray-100 hover:text-zinc-900",
    )

    const style: React.CSSProperties = {
        top,
        left,
        height: tileHeight,
        fontSize: 8,
        "--tile-x": `${left}px`,
        "--tile-y": `${top}px`,
    } as React.CSSProperties

    return (
        <div className={cn(className)} style={style}>
            {tile.coordinate?.row}, {tile.coordinate?.column}
        </div>
    )
}

export default TileView
