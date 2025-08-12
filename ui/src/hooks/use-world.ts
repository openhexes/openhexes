import { create } from "@bufbuild/protobuf"
import { GetSampleWorldRequestSchema } from "proto/ts/game/v1/game_pb"
import { type Tile } from "proto/ts/map/v1/tile_pb"
import type { Progress } from "proto/ts/progress/v1/progress_pb"
import { type World, WorldSchema } from "proto/ts/world/v1/world_pb"
import React from "react"

import { GameClient } from "./fetch"

const sleep = async (ms: number) => {
    return await new Promise((resolve) => setTimeout(resolve, ms))
}

type S = {
    world: World
    selectedDepth: number
    selectDepth: (depth: number) => void
    selectedTile?: Tile
    selectTile: (tile: Tile) => void
    smoothPan?: boolean
    setSmoothPan?: (smooth: boolean) => void
    useDetailedSvg?: boolean
    setUseDetailedSvg?: (detailed: boolean) => void
}

export const WorldContext = React.createContext<S>({
    world: create(WorldSchema),
    selectedDepth: 0,
    selectDepth: () => null,
    selectTile: () => null,
})

export const useWorld = () => {
    const context = React.useContext(WorldContext)
    if (context === undefined) throw new Error("useWorld must be used within a GridView")
    return context
}

const buildWorld = async (
    totalRows: number,
    totalColumns: number,
    maxRowsPerSegment: number,
    maxColumnsPerSegment: number,
    setProgress: React.Dispatch<React.SetStateAction<Progress | undefined>>,
): Promise<World> => {
    const request = create(GetSampleWorldRequestSchema, {
        totalLayers: 5,
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

        for (const chunk of response.world.layers) {
            const layer = world.layers[chunk.depth]
            if (layer === undefined) {
                world.layers[chunk.depth] = chunk
            } else {
                layer.segmentRows.push(...chunk.segmentRows)
            }
        }
    }

    await sleep(250)

    console.info("complete world", world)

    return world
}

export const useFetchedWorld = (
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
