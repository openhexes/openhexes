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
    const [selected, setSelected] = React.useState<number>(0)

    const { row, column } = lib.getCoordinates(tile)
    const even = row % 2 === 0
    const top = row * height * 0.75
    const left = column * width + (even ? 0 : width / 2)

    const className = cn(
        "tile",
        "select-none flex items-center justify-center absolute",
        "text-xs text-transparent bg-transparent hover:bg-gray-100 hover:text-zinc-900 opacity-50",
        { "bg-sky-400 text-zinc-900": selected === 2 },
        { "bg-rose-400 text-zinc-900": selected === 1 },
    )

    const handleClick = (e: React.MouseEvent) => {
        if (selected) {
            setSelected(0)
            return
        }

        if (e.shiftKey) {
            setSelected(2)
        } else {
            setSelected(1)
        }
    }

    const style: React.CSSProperties = {
        top,
        left,
        height,
        width,
        fontSize: 8,
    } as React.CSSProperties

    return (
        <div className={cn(className)} style={style} onClick={handleClick}>
            <div>
                {tile.coordinate?.row}, {tile.coordinate?.column}
            </div>
        </div>
    )
}

export default TileView
