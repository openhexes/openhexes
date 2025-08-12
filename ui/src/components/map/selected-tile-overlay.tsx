import * as tileUtil from "@/lib/tiles"
import type { Tile } from "proto/ts/map/v1/tile_pb"

interface SelectedTileOverlayProps {
    tile: Tile | undefined
    tileWidth: number
    tileHeight: number
    rowHeight: number
}

export function SelectedTileOverlay({
    tile,
    tileWidth,
    tileHeight,
    rowHeight,
}: SelectedTileOverlayProps) {
    if (!tile) return null

    const { row, column } = tileUtil.getCoordinates(tile)
    const even = row % 2 === 0
    const left = column * tileWidth + (even ? 0 : tileWidth / 2)
    const top = row * rowHeight

    return (
        <div
            className="absolute pointer-events-none bg-lime-200"
            style={{
                left,
                top,
                width: tileWidth,
                height: tileHeight,
                zIndex: 999,
                clipPath: "var(--shape-hex)",
            }}
        ></div>
    )
}
