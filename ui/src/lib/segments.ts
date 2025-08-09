import type { Segment } from "proto/ts/map/v1/tile_pb"

export const getBounds = (segment: Segment) => {
    const { minRow = 0, maxRow = 0, minColumn = 0, maxColumn = 0 } = segment.bounds ?? {}
    return { minRow, maxRow, minColumn, maxColumn }
}

export const getKey = (segment: Segment): string => {
    const { minRow, maxRow, minColumn, maxColumn } = getBounds(segment)
    return `[${minRow},${maxRow}),[${minColumn},${maxColumn})`
}

export const getDimensions = (segment: Segment, tileWidth: number, tileHeight: number) => {
    const { minRow, maxRow, minColumn, maxColumn } = getBounds(segment)
    const width = (maxColumn - minColumn + 1) * tileWidth
    const height = (maxRow - minRow + 1) * tileHeight * 0.75 + tileHeight * 0.25
    return { width, height }
}

// Given a segment's bounds, calculate its top-left pixel coordinates
export function getSegmentOriginPx(segment: Segment, tileWidth: number, tileHeight: number) {
    const { minRow, minColumn } = getBounds(segment)

    // Each row in a pointy-top hex grid is offset by 0.75 * tileHeight vertically
    const rowHeight = tileHeight * 0.75

    // Hexes in odd rows are shifted horizontally by half a tile width
    const columnShift = minRow % 2 === 0 ? 0 : tileWidth / 2

    const originX = minColumn * tileWidth + columnShift
    const originY = minRow * rowHeight

    return { left: originX, top: originY }
}
