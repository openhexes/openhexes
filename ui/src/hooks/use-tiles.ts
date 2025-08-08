import { create } from "@bufbuild/protobuf"
import { GetSampleGridRequestSchema } from "proto/ts/game/v1/game_pb"
import { type Grid, GridSchema } from "proto/ts/map/v1/tile_pb"
import type { Progress } from "proto/ts/progress/v1/progress_pb"
import React from "react"

import { GameClient } from "./fetch"

export const useTileDimensions = () => {
    const tileHeight = 60 // long diagonal
    const tileWidth = (Math.sqrt(3) * tileHeight) / 2 // short diagonal

    const sideLength = tileHeight / 2
    const triangleHeight = Math.sqrt(4 * sideLength ** 2 - tileWidth ** 2) / 2
    const rowHeight = tileHeight - triangleHeight

    return { tileHeight, tileWidth, rowHeight, triangleHeight }
}

const sleep = async (ms: number) => {
    return await new Promise((resolve) => setTimeout(resolve, ms))
}

const buildTileGrid = async (
    totalRows: number,
    totalColumns: number,
    maxRowsPerSegment: number,
    maxColumnsPerSegment: number,
    setProgress: React.Dispatch<React.SetStateAction<Progress | undefined>>,
): Promise<Grid> => {
    const request = create(GetSampleGridRequestSchema, {
        totalRows,
        totalColumns,
        maxRowsPerSegment,
        maxColumnsPerSegment,
    })

    const grid = create(GridSchema)

    for await (const response of GameClient.getSampleGrid(request, { timeoutMs: 60000 })) {
        console.info(response)
        if (response.progress) {
            setProgress(response.progress)
        }
        if (response.grid) {
            if (response.grid.totalRows > 0) {
                grid.totalRows = response.grid.totalRows
            }
            if (response.grid.totalColumns > 0) {
                grid.totalColumns = response.grid.totalColumns
            }
            if (response.grid.segmentRows) {
                grid.segmentRows.push(...response.grid.segmentRows)
            }
        }
    }

    await sleep(250)

    return grid
}

export const useTileGrid = (
    totalRows: number,
    totalColumns: number,
    maxRowsPerSegment = 15,
    maxColumnsPerSegment = 15,
) => {
    const [progress, setProgress] = React.useState<Progress | undefined>(undefined)
    const [isLoading, setIsLoading] = React.useState<boolean>(true)
    const [error, setError] = React.useState<Error | undefined>(undefined)
    const [grid, setGrid] = React.useState<Grid | undefined>(undefined)

    const promise = buildTileGrid(
        totalRows,
        totalColumns,
        maxRowsPerSegment,
        maxColumnsPerSegment,
        setProgress,
    )
        .then(setGrid)
        .catch(setError)
        .finally(() => setIsLoading(false))

    return { promise, isLoading, grid, error, progress }
}
