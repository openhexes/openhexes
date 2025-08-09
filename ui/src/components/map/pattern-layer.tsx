import { segmentOriginWorldPx } from "@/lib/segments"
import type { Segment } from "proto/ts/map/v1/tile_pb"
import React from "react"

interface P {
    segment: Segment
    tileWidth: number
    tileHeight: number
}

export const PatternLayer: React.FC<P> = ({ segment, tileHeight, tileWidth }) => {
    const svgMarkup = segment.renderingSpec?.svg || ""
    if (!svgMarkup) return null

    const { x, y } = segmentOriginWorldPx(segment, tileWidth, tileHeight)

    // Optional: if you want the wrapper sized exactly to the segment
    const minR = segment.bounds?.minRow ?? 0
    const maxR = segment.bounds?.maxRow ?? 0
    const minC = segment.bounds?.minColumn ?? 0
    const maxC = segment.bounds?.maxColumn ?? 0
    const rowHeight = 0.75 * tileHeight
    const segWidth =
        (maxC + 1) * tileWidth +
        (maxR % 2 !== 0 ? tileWidth / 2 : 0) -
        (minC * tileWidth + (minR % 2 !== 0 ? tileWidth / 2 : 0))
    const segHeight = maxR * rowHeight + tileHeight - minR * rowHeight

    return (
        <div
            className="absolute pointer-events-none"
            style={{ left: x, top: y, width: segWidth, height: segHeight }}
            dangerouslySetInnerHTML={{ __html: svgMarkup }}
        />
    )
}
