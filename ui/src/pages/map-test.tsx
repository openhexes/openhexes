import { ErrorView } from "@/components/utils/error"
import { ProgressView } from "@/components/utils/progress-view"
import { useWorld } from "@/hooks/use-tiles"
import React from "react"

const Map = React.lazy(() => import("@/components/map/grid-view"))

const rowCount = 300
const columnCount = 300

export const MapTest = () => {
    const { world, isLoading, progress, error } = useWorld(rowCount, columnCount, 30, 30)

    if (isLoading) {
        return <ProgressView progress={progress} />
    }

    if (world?.layers[0] !== undefined) {
        return <Map grid={world.layers[0]} />
    }

    return <ErrorView error={error ?? new Error("unknown error")} />
}

export default MapTest
