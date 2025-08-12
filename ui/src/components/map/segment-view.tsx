import { segmentOriginWorldPx } from "@/lib/segments"
import type { Segment } from "proto/ts/map/v1/tile_pb"
import React from "react"

interface SegmentViewProps {
    segment: Segment
    tileWidth: number
    tileHeight: number
    useDetailedSvg?: boolean
}

export const SegmentView: React.FC<SegmentViewProps> = ({
    segment,
    tileHeight,
    tileWidth,
    useDetailedSvg = false,
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

    // Use SVG data based on detail level preference
    const renderingSpec = segment.renderingSpec
    const svgContent = useDetailedSvg ? renderingSpec?.svg : renderingSpec?.svgLightweight

    if (!svgContent) {
        return null
    }

    return (
        <div
            className="absolute pointer-events-none"
            style={{
                left: x,
                top: y,
                width: segWidth,
                height: segHeight,
            }}
            dangerouslySetInnerHTML={{ __html: svgContent }}
        />
    )
}
