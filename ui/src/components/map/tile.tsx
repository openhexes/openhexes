import * as util from "@/lib/tiles"
import { cn } from "@/lib/utils"
import React from "react"

import "./tile.css"

export interface TileProps {
    tile: util.AnnotatedTile
}

export const Tile: React.FC<TileProps> = ({ tile }) => {
    return (
        <div
            className={cn(tile.className, "bg-gray-950 text-transparent hover:text-gray-500")}
            style={{ top: tile.top, left: tile.left, height: tile.height }}
        >
            {tile.proto.coordinate?.row}, {tile.proto.coordinate?.column}
        </div>
    )
}

export default Tile
