import { useTileDimensions } from "@/hooks/use-tiles"
import * as lib from "@/lib/tiles"
import { cn } from "@/lib/utils"
import { Direction } from "proto/ts/map/v1/compass_pb"
import type { Tile } from "proto/ts/map/v1/tile_pb"
import React from "react"

import "./tile.css"

export interface TileProps {
    tile: Tile
}

export const TileView: React.FC<TileProps> = ({ tile }) => {
    const { tileHeight, tileWidth } = useTileDimensions()

    const { row, column } = lib.getCoordinates(tile)
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
        width: tileWidth,
        fontSize: 8,
    } as React.CSSProperties

    const verts = lib.hexVerts(tileWidth, tileHeight)
    const inner = lib.insetVerts(verts, 0.84) // 0.80..0.90; lower = thicker
    const edges = tile.renderingSpec?.edges ?? []
    const edgesSVG = (
        <svg
            className="absolute inset-0 pointer-events-none z-10"
            width={tileWidth}
            height={tileHeight}
            viewBox={`0 0 ${tileWidth} ${tileHeight}`}
            aria-hidden
        >
            {edges.map((e, i) => {
                if (e.direction === Direction.UNSPECIFIED) {
                    return null
                }

                const pair = lib.SEG_BY_DIR[e.direction]
                if (!pair) return null
                const [iA, iB] = pair
                const A = verts[iA],
                    B = verts[iB],
                    Ai = inner[iA],
                    Bi = inner[iB]

                const paint = lib.edgePaint(tile.terrainId || "", e.neighbourTerrainId || "")

                // underlay (shadow/rim), slightly larger to create soft edge and kill seams
                return (
                    <g key={i}>
                        {paint.under && (
                            <path
                                d={lib.polyD(A, B, Bi, Ai, paint.under.grow ?? 0.5)}
                                fill={paint.under.fill}
                                shapeRendering="geometricPrecision"
                            />
                        )}
                        <path
                            d={lib.polyD(A, B, Bi, Ai)}
                            fill={paint.fill}
                            shapeRendering="geometricPrecision"
                        />
                        {/* optional rounded caps to smooth junctions */}
                        <circle cx={A.x} cy={A.y} r={1.0} fill={paint.fill} />
                        <circle cx={B.x} cy={B.y} r={1.0} fill={paint.fill} />
                    </g>
                )
            })}
        </svg>
    )

    return (
        <div className={cn(className)} style={style}>
            <div>
                {tile.coordinate?.row}, {tile.coordinate?.column}
            </div>

            {edgesSVG}
        </div>
    )
}

export default TileView
