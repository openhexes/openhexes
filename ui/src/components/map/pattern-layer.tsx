import type { Segment } from "proto/ts/map/v1/tile_pb"
import React from "react"

interface PatternLayerProps {
    segment: Segment
}

export const PatternLayer: React.FC<PatternLayerProps> = ({ segment }) => {
    const svgMarkup = segment.renderingSpec?.svg || ""
    if (!svgMarkup) return null

    return (
        <div
            className="absolute pointer-events-none"
            style={{ left: 0, top: 0 }} // ← always 0,0
            dangerouslySetInnerHTML={{ __html: svgMarkup }} // SVG’s viewBox positions content
        />
    )
}
