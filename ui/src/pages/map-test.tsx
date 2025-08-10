import { ErrorView } from "@/components/utils/error"
import { ProgressView } from "@/components/utils/progress-view"
import { useWorld } from "@/hooks/use-tiles"
import { useWindowSize } from "@uidotdev/usehooks"
import React from "react"

const Map = React.lazy(() => import("@/components/map/grid-view"))

const rowCount = 100
const columnCount = 100

export const MapTest = () => {
    const { height, width } = useWindowSize()
    const { world, isLoading, progress, error } = useWorld(rowCount, columnCount, 16, 20)

    if (isLoading) {
        return <ProgressView progress={progress} />
    }

    if (world?.layers[0] !== undefined) {
        return (
            <Map height={height || 600} width={width || 800} world={world} grid={world.layers[0]} />
        )
    }

    return <ErrorView error={error ?? new Error("unknown error")} />
}

export default MapTest
