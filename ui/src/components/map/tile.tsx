import { useTiles } from "@/hooks/use-tiles"
import { cn } from "@/lib/utils"
import React from "react"

import "./tile.css"

export interface TileProps {
    row: number
    column: number
    visible: boolean
}

export const Tile: React.FC<TileProps> = ({ row, column, visible }) => {
    const ref = React.useRef<HTMLDivElement>(null)

    const { height, width } = useTiles()

    const evenRow = row % 2 === 0
    const top = row * height * 0.75
    const left = column * width + (evenRow ? 0 : width / 2)

    return (
        <div
            ref={ref}
            className={cn(
                "tile",
                "select-none flex items-center justify-center absolute",
                "text-xs bg-background text-transparent hover:bg-gray-900 hover:text-gray-600",
                { "text-rose-700": !visible },
            )}
            style={{ top, left, height }}
        >
            {row}, {column}
        </div>
    )
}

export default Tile
