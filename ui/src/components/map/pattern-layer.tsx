import { getCoordinates } from "@/lib/tiles"
import { type Tile as PTile } from "proto/ts/map/v1/tile_pb"
import React from "react"

type Props = {
    id: string // e.g., "water"
    tiles: PTile[] // pass visible tiles
    filter: (t: PTile) => boolean
    tileWidth: number
    tileHeight: number
    mapWidth: number
    mapHeight: number
    // pattern tile size in px (controls hatch/dot spacing)
    cell?: number
    // pattern SVG content (tiny string)
    svgTile: string
    opacity?: number
}

/** Pure SVG pattern layer that moves with the map (no CSS writes on pan). */
export const PatternLayer: React.FC<Props> = ({
    id,
    tiles,
    filter,
    tileWidth,
    tileHeight,
    mapWidth,
    mapHeight,
    cell = 16,
    svgTile,
    opacity = 1,
}) => {
    const patternId = `pat-${id}`
    const maskId = `mask-${id}`

    return (
        <svg
            width={mapWidth}
            height={mapHeight}
            style={{
                position: "absolute",
                left: 0,
                top: 0,
                pointerEvents: "none",
                opacity,
            }}
            aria-hidden
        >
            <defs>
                {/* Tiled pattern in world/user space, using a tiny inline SVG tile */}
                <pattern id={patternId} width={cell} height={cell} patternUnits="userSpaceOnUse">
                    <image
                        href={`data:image/svg+xml;utf8,${encodeURIComponent(svgTile)}`}
                        width={cell}
                        height={cell}
                        x="0"
                        y="0"
                        preserveAspectRatio="none"
                    />
                </pattern>

                {/* Union-of-tiles mask */}
                <mask id={maskId} maskUnits="userSpaceOnUse" maskContentUnits="userSpaceOnUse">
                    {/* Black clears; white shows */}
                    <rect x={-1e6} y={-1e6} width={2e6} height={2e6} fill="black" />
                    <g fill="white">
                        {tiles.filter(filter).map((t) => {
                            const d = hexPathD(t, tileWidth, tileHeight)
                            const { row, column, depth } = getCoordinates(t)
                            return <path key={`${id}:${row},${column},${depth}`} d={d} />
                        })}
                    </g>
                </mask>
            </defs>

            {/* Full-map rect that gets patterned + masked */}
            <rect
                x="0"
                y="0"
                width={mapWidth}
                height={mapHeight}
                fill={`url(#${patternId})`}
                mask={`url(#${maskId})`}
            />
        </svg>
    )
}

function hexPathD(t: PTile, tileWidth: number, tileHeight: number): string {
    const { row, column } = getCoordinates(t)

    // same placement math you use in <TileView>
    const even = row % 2 === 0
    const x = column * tileWidth + (even ? 0 : tileWidth / 2)
    const y = row * tileHeight * 0.75

    // pointy-top hex: top/bottom are points, left/right are flat
    const w = tileWidth
    const h = tileHeight
    const v = h / 4 // vertical inset of the “shoulders”

    const p1 = [x + w / 2, y] // top point
    const p2 = [x + w, y + v] // upper-right
    const p3 = [x + w, y + 3 * v] // lower-right
    const p4 = [x + w / 2, y + h] // bottom point
    const p5 = [x, y + 3 * v] // lower-left
    const p6 = [x, y + v] // upper-left

    // eslint-disable-next-line @typescript-eslint/restrict-template-expressions
    return `M${p1} L${p2} L${p3} L${p4} L${p5} L${p6} Z`
}
