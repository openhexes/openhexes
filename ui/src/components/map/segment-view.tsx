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

export const SegmentView: React.FC<PatternLayerProps> = ({
    segment,
    tileHeight,
    tileWidth,
    isZoomedOut,
}) => {
    const { x, y } = segmentOriginWorldPx(segment, tileWidth, tileHeight)

    // Calculate segment dimensions
    const minR = segment.bounds?.minRow ?? 0
    const maxR = segment.bounds?.maxRow ?? 0
    const minC = segment.bounds?.minColumn ?? 0
    const maxC = segment.bounds?.maxColumn ?? 0
    const rowHeight = 0.75 * tileHeight
    const segWidth = (maxC - minC + 1) * tileWidth + tileWidth / 2 // Add padding for offset rows
    const segHeight = (maxR - minR + 1) * rowHeight + tileHeight * 0.25 // Add padding for tile height

    // Use WebP data based on zoom level
    const renderingSpec = segment.renderingSpec
    const webpData = isZoomedOut ? renderingSpec?.webpLightweight : renderingSpec?.webp

    // WebP should always be available - no fallbacks needed
    if (!webpData || webpData.length === 0) {
        console.error(`Missing WebP data for segment ${segment.bounds?.depth}.${minR}.${minC}`)
        return null
    }

    // Convert bytes to blob URL (WebP format)
    const blob = new Blob([new Uint8Array(webpData)], { type: "image/webp" })
    const imageUrl = URL.createObjectURL(blob)

    return (
        <img
            src={imageUrl}
            alt={`Segment ${segment.bounds?.depth}.${minR}.${minC}`}
            className="absolute pointer-events-none"
            style={{
                left: x,
                top: y,
                width: segWidth,
                height: segHeight,
                imageRendering: "pixelated", // Crisp scaling for pixel art
            }}
            onLoad={() => URL.revokeObjectURL(imageUrl)} // Clean up blob URL
        />
    )
}
