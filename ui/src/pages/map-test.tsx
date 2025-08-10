import { ErrorView } from "@/components/utils/error"
import { ProgressView } from "@/components/utils/progress-view"
import { useFetchedWorld } from "@/hooks/use-world"
import { useWindowSize } from "@uidotdev/usehooks"
import React from "react"

const Map = React.lazy(() => import("@/components/map/grid-view"))

const rowCount = 64
const columnCount = 64

export const MapTest = () => {
    const { height, width } = useWindowSize()
    const { world, isLoading, progress, error } = useFetchedWorld(rowCount, columnCount, 16, 20)

    if (!height || !width) {
        return null
    }

    if (isLoading) {
        return <ProgressView progress={progress} />
    }

    if (world?.layers[0] !== undefined) {
        return <Map height={height} width={width} world={world} grid={world.layers[0]} />
    }

    return <ErrorView error={error ?? new Error("unknown error")} />
}

export default MapTest
