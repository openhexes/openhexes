// src/components/pattern-layer.tsx
import { getCoordinates } from "@/lib/tiles"
import { type Tile as PTile } from "proto/ts/map/v1/tile_pb"
import React from "react"

type Props = {
    id: string // e.g. "water"
    tiles: PTile[] // visible tiles (same set you render)
    filter: (t: PTile) => boolean
    tileWidth: number
    tileHeight: number
    mapWidth: number
    mapHeight: number
    className: string // e.g. "pattern--water"
}

/**
 * One world-sized layer with a repeating background.
 * It's masked by the union of tiles that pass `filter`.
 * IMPORTANT: Put this inside the same translated container as your tiles.
 */
export const PatternLayer: React.FC<Props> = ({
    id,
    tiles,
    filter,
    tileWidth,
    tileHeight,
    mapWidth,
    mapHeight,
    className,
}) => {
    const filteredTiles = tiles.filter(filter).map((t) => {
        const d = hexPathD(t, tileWidth, tileHeight)
        // key is cheap + stable
        const { row, column, depth } = getCoordinates(t)
        return <path key={`${id}:${row},${column},${depth}`} d={d} />
    })
    console.info("pattern", id, filteredTiles)

    return (
        <>
            {/* Mask defs must be in the DOM */}
            <svg width={0} height={0} aria-hidden>
                <defs>
                    <mask
                        id={`mask-${id}`}
                        maskUnits="userSpaceOnUse"
                        maskContentUnits="userSpaceOnUse"
                    >
                        {/* Black clears everything; white shows pattern */}
                        <rect x={-1e6} y={-1e6} width={2e6} height={2e6} fill="black" />
                        <g fill="white">{filteredTiles}</g>
                    </mask>
                </defs>
            </svg>

            {/* The pattern itself; masked by the union above */}
            <div
                className={`pattern ${className}`}
                style={{
                    position: "absolute",
                    left: 0,
                    top: 0,
                    width: mapWidth,
                    height: mapHeight,
                    // Critical: mask applied to the world-sized layer
                    mask: `url(#mask-${id})`,
                    WebkitMask: `url(#mask-${id})`,
                    pointerEvents: "none",
                }}
                aria-hidden
            />
        </>
    )
}

/** Hex path in world pixels that matches your TileView placement */
function hexPathD(t: PTile, tileWidth: number, tileHeight: number): string {
    const { row, column } = getCoordinates(t)
    const even = row % 2 === 0
    const x = column * tileWidth + (even ? 0 : tileWidth / 2)
    const y = row * tileHeight * 0.75

    // tweak these two if your hex proportions differ
    const a = tileWidth / 4 // horizontal corner inset
    const b = tileHeight / 4 // vertical corner inset

    const p1 = [x + a, y]
    const p2 = [x + tileWidth - a, y]
    const p3 = [x + tileWidth, y + b]
    const p4 = [x + tileWidth - a, y + tileHeight]
    const p5 = [x + a, y + tileHeight]
    const p6 = [x, y + b]

    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    return `M${p1} L${p2} L${p3} L${p4} L${p5} L${p6} Z`
}
