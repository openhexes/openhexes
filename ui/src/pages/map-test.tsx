import { ErrorView } from "@/components/utils/error"
import { ProgressView } from "@/components/utils/progress-view"
import { useTileGrid } from "@/hooks/use-tiles"
import React from "react"

const Map = React.lazy(() => import("@/components/map/grid-view"))

const rowCount = 300
const columnCount = 300

export const MapTest = () => {
    const { grid, isLoading, progress, error } = useTileGrid(rowCount, columnCount, 30, 30)

    if (isLoading) {
        return <ProgressView progress={progress} />
    }

    if (grid !== undefined) {
        return <Map grid={grid} />
    }

    return <ErrorView error={error ?? new Error("unknown error")} />
}

export default MapTest
