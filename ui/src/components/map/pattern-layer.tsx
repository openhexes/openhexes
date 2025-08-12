import { segmentOriginWorldPx } from "@/lib/segments"
import type { Segment } from "proto/ts/map/v1/tile_pb"
import React from "react"

interface P {
    segment: Segment
    tileWidth: number
    tileHeight: number
}

interface PatternLayerProps extends P {
    isZoomedOut?: boolean
}

export const PatternLayer: React.FC<PatternLayerProps> = ({ segment, tileHeight, tileWidth }) => {
    // TESTING: Remove SVGs completely, just show labels + borders to test panning performance
    const { x, y } = segmentOriginWorldPx(segment, tileWidth, tileHeight)
    
    // Calculate segment dimensions
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
            className="absolute pointer-events-none border-2 border-red-500 bg-blue-100 bg-opacity-20 flex items-center justify-center"
            style={{ 
                left: x, 
                top: y, 
                width: segWidth, 
                height: segHeight,
                fontSize: '24px',
                fontWeight: 'bold',
                color: 'red',
            }}
        >
            Segment {minR}-{minC}
        </div>
    )
}
