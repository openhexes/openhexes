import { type Grid, type Segment, annotate, boundsInclude } from "@/lib/tiles"
import type { Tile } from "proto/ts/map/v1/tile_pb"
import React from "react"

export const useTileDimensions = () => {
    const tileHeight = 60 // long diagonal
    const tileWidth = (Math.sqrt(3) * tileHeight) / 2 // short diagonal

    const sideLength = tileHeight / 2
    const triangleHeight = Math.sqrt(4 * sideLength ** 2 - tileWidth ** 2) / 2
    const rowHeight = tileHeight - triangleHeight

    return { tileHeight, tileWidth, rowHeight, triangleHeight }
}

const animationFrame = async () => {
    return await new Promise((resolve) => requestAnimationFrame(resolve))
}

const sleep = async (ms: number) => {
    return await new Promise((resolve) => setTimeout(resolve, ms))
}

interface ProgressSegment {
    state: "waiting" | "running" | "done"
    title: string
    subtitle?: string
    durationMs?: number
}

const buildTileGrid = async (
    tiles: Tile[],
    totalRows: number,
    totalColumns: number,
    tileHeight: number,
    tileWidth: number,
    maxRowsPerSegment: number,
    maxColumnsPerSegment: number,
    setProgress: React.Dispatch<React.SetStateAction<number>>,
    setProgressSegments: React.Dispatch<React.SetStateAction<ProgressSegment[]>>,
): Promise<Grid> => {
    const psPrepareContainers: ProgressSegment = {
        state: "running",
        title: "Prepare segment containers",
    }
    const psArrangeSegments: ProgressSegment = {
        state: "waiting",
        title: "Arrange grid",
    }
    const psProcessTiles: ProgressSegment = {
        state: "waiting",
        title: "Process tiles",
    }
    const updateProgressSegments = () => {
        setProgressSegments([psPrepareContainers, psArrangeSegments, psProcessTiles])
    }
    updateProgressSegments()

    // prepare segment containers
    let start = performance.now()
    const segments: Segment[] = []
    for (let rowStart = 0; rowStart < totalRows; rowStart += maxRowsPerSegment) {
        await animationFrame()

        for (let columnStart = 0; columnStart < totalColumns; columnStart += maxColumnsPerSegment) {
            const rowEnd = rowStart + maxRowsPerSegment
            const columnEnd = columnStart + maxColumnsPerSegment
            segments.push({
                tiles: [],
                bounds: {
                    minRow: rowStart,
                    maxRow: rowEnd,
                    minColumn: columnStart,
                    maxColumn: columnEnd,
                },
            })
        }
    }
    setProgress(0.1)

    psPrepareContainers.durationMs = performance.now() - start
    psPrepareContainers.state = "done"
    psArrangeSegments.state = "running"
    updateProgressSegments()

    // arrange segments in a grid
    start = performance.now()
    const segmentRows: Segment[][] = []
    let gridRow: Segment[] = []
    let rowStart: number | undefined = undefined

    const segmentProgressCost = 0.2 / segments.length
    for (let i = 0; i < segments.length; i++) {
        await animationFrame()

        const segment = segments[i]

        if (rowStart === undefined) {
            rowStart = segment.bounds.minRow
            gridRow = [segment]
        } else if (rowStart !== segment.bounds.minRow) {
            segmentRows.push(gridRow)
            gridRow = [segment]
            rowStart = segment.bounds.minRow
        } else {
            gridRow.push(segment)
        }

        setProgress((v) => v + segmentProgressCost)
    }
    segmentRows.push(gridRow)

    psArrangeSegments.durationMs = performance.now() - start
    psArrangeSegments.state = "done"
    psProcessTiles.state = "running"
    updateProgressSegments()

    const tileProgressCost = 0.7 / tiles.length

    // put tiles into respective segments
    start = performance.now()

    // todo: segment search can be optimized
    let processedTileCount = 0
    for (const tile of tiles) {
        if (processedTileCount % 10000 === 0) {
            await animationFrame()

            psProcessTiles.subtitle = `${processedTileCount} / ${tiles.length}`
            updateProgressSegments()

            setProgress(0.3 + processedTileCount * tileProgressCost)
        }

        for (const segmentRow of segmentRows) {
            for (const segment of segmentRow) {
                if (boundsInclude(tile, segment.bounds)) {
                    segment.tiles.push(annotate(tile, tileHeight, tileWidth))
                }
            }
        }

        processedTileCount++
    }

    psProcessTiles.durationMs = performance.now() - start
    psProcessTiles.state = "done"
    updateProgressSegments()

    setProgress(1)
    await sleep(1000)

    return { segmentRows, totalRows, totalColumns }
}

export const useTileGrid = (
    tiles: Tile[],
    totalRows: number,
    totalColumns: number,
    maxRowsPerSegment = 15,
    maxColumnsPerSegment = 15,
) => {
    const { tileHeight, tileWidth } = useTileDimensions()

    const [progress, setProgress] = React.useState<number>(0)
    const [progressSegments, setProgressSegments] = React.useState<ProgressSegment[]>([])
    const [isLoading, setIsLoading] = React.useState<boolean>(true)
    const [error, setError] = React.useState<Error | undefined>(undefined)
    const [grid, setGrid] = React.useState<Grid | undefined>(undefined)

    const promise = buildTileGrid(
        tiles,
        totalRows,
        totalColumns,
        tileHeight,
        tileWidth,
        maxRowsPerSegment,
        maxColumnsPerSegment,
        setProgress,
        setProgressSegments,
    )
        .then(setGrid)
        .catch(setError)
        .finally(() => setIsLoading(false))

    return { promise, isLoading, grid, error, progress, progressSegments }
}
