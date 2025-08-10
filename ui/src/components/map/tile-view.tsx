import { useWorld } from "@/hooks/use-world"
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
    const { selectedTile, selectTile } = useWorld()

    const { row, column } = lib.getCoordinates(tile)
    const even = row % 2 === 0
    const top = row * height * 0.75
    const left = column * width + (even ? 0 : width / 2)

    const selected = selectedTile?.key === tile.key

    const className = cn(
        "tile",
        "select-none flex items-center justify-center absolute",
        "text-s text-transparent bg-transparent hover:text-black opacity-20",
    )

    const style: React.CSSProperties = {
        top,
        left,
        height,
        width,
        fontSize: 8,
    } as React.CSSProperties

    return (
        <div
            className={cn(className, { "bg-gray-100": selected })}
            onMouseEnter={() => selectTile(tile)}
            onClick={() => selectTile(tile)}
            style={style}
        />
    )
}

export default TileView
