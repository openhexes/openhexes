import { create } from "@bufbuild/protobuf"
import { GetSampleWorldRequestSchema } from "proto/ts/game/v1/game_pb"
import { GridSchema } from "proto/ts/map/v1/tile_pb"
import type { Progress } from "proto/ts/progress/v1/progress_pb"
import { type World, WorldSchema } from "proto/ts/world/v1/world_pb"
import React from "react"

import { GameClient } from "./fetch"

const sleep = async (ms: number) => {
    return await new Promise((resolve) => setTimeout(resolve, ms))
}

const buildWorld = async (
    totalRows: number,
    totalColumns: number,
    maxRowsPerSegment: number,
    maxColumnsPerSegment: number,
    setProgress: React.Dispatch<React.SetStateAction<Progress | undefined>>,
): Promise<World> => {
    const request = create(GetSampleWorldRequestSchema, {
        totalRows,
        totalColumns,
        maxRowsPerSegment,
        maxColumnsPerSegment,
    })

    const world = create(WorldSchema, {
        terrainRegistry: {},
        spellRegistry: {},
        creatureRegistry: {},
    })

    for await (const response of GameClient.getSampleWorld(request, { timeoutMs: 60000 })) {
        console.debug(response)
        if (response.progress) {
            setProgress(response.progress)
        }

        if (!response.world) {
            continue
        }

        if (response.world.renderingSpec) {
            world.renderingSpec = response.world.renderingSpec
        }

        for (const [k, v] of Object.entries(response.world.terrainRegistry)) {
            world.terrainRegistry[k] = v
        }
        for (const [k, v] of Object.entries(response.world.spellRegistry)) {
            world.spellRegistry[k] = v
        }
        for (const [k, v] of Object.entries(response.world.creatureRegistry)) {
            world.creatureRegistry[k] = v
        }

        for (const [idx, incoming] of response.world.layers.entries()) {
            if (world.layers[idx] === undefined) {
                world.layers[idx] = create(GridSchema)
            }

            const grid = world.layers[idx]
            if (incoming.totalRows > 0) {
                grid.totalRows = incoming.totalRows
            }
            if (incoming.totalColumns > 0) {
                grid.totalColumns = incoming.totalColumns
            }
            if (incoming.segmentRows) {
                grid.segmentRows.push(...incoming.segmentRows)
            }
        }
    }

    await sleep(250)

    const findTile = (row: number, col: number) => {
        const grid = world.layers[0]
        for (const segmentRow of grid.segmentRows) {
            for (const segment of segmentRow.segments) {
                for (const tile of segment.tiles) {
                    if (tile.coordinate?.row === row && tile.coordinate?.column === col) {
                        return tile
                    }
                }
            }
        }
    }

    // eslint-disable-next-line
    const win = window as any
    // eslint-disable-next-line
    win.findTile = findTile

    console.info("complete world", world)

    return world
}

export const useWorld = (
    totalRows: number,
    totalColumns: number,
    maxRowsPerSegment = 15,
    maxColumnsPerSegment = 15,
) => {
    const [progress, setProgress] = React.useState<Progress | undefined>(undefined)
    const [isLoading, setIsLoading] = React.useState<boolean>(true)
    const [error, setError] = React.useState<Error | undefined>(undefined)
    const [world, setWorld] = React.useState<World | undefined>(undefined)

    const promise = buildWorld(
        totalRows,
        totalColumns,
        maxRowsPerSegment,
        maxColumnsPerSegment,
        setProgress,
    )
        .then(setWorld)
        .catch(setError)
        .finally(() => setIsLoading(false))

    return { promise, isLoading, world, error, progress }
}
