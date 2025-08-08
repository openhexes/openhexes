import { useTileDimensions } from "@/hooks/use-tiles"
import { cn } from "@/lib/utils"
import type { Tile } from "proto/ts/map/v1/tile_pb"
import React from "react"

import "./tile.css"

export interface TileProps {
    tile: Tile
}

export const TileView: React.FC<TileProps> = ({ tile }) => {
    const { tileHeight } = useTileDimensions()

    return (
        <div
            className={cn(
                tile.renderingSpec?.className,
                "bg-gray-950 text-transparent hover:text-gray-500",
            )}
            style={{
                top: tile.renderingSpec?.top,
                left: tile.renderingSpec?.left,
                height: tileHeight,
            }}
        >
            {tile.coordinate?.row}, {tile.coordinate?.column}
        </div>
    )
}

export default TileView
