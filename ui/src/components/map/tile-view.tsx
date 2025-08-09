import * as lib from "@/lib/tiles"
import { cn } from "@/lib/utils"
import type { Tile } from "proto/ts/map/v1/tile_pb"
import React from "react"

import "./tile.css"

export interface TileProps {
    tile: Tile
    height: number
    width: number
}

export const TileView: React.FC<TileProps> = ({ tile, height, width }) => {
    const { row, column } = lib.getCoordinates(tile)
    const even = row % 2 === 0
    const top = row * height * 0.75
    const left = column * width + (even ? 0 : width / 2)

    const className = cn(
        "tile",
        "select-none flex items-center justify-center absolute",
        "text-xs text-transparent bg-transparent hover:bg-gray-100 hover:text-zinc-900",
    )

    const style: React.CSSProperties = {
        top,
        left,
        height,
        width,
        fontSize: 8,
    } as React.CSSProperties

    return (
        <div className={cn(className)} style={style}>
            <div>
                {tile.coordinate?.row}, {tile.coordinate?.column}
            </div>
        </div>
    )
}

export default TileView
