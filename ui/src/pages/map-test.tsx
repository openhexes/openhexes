import { ErrorView } from "@/components/utils/error-view"
import { ProgressView } from "@/components/utils/progress-view"
import { useFetchedWorld } from "@/hooks/use-world"
import { useWindowSize } from "@uidotdev/usehooks"
import React from "react"

const Map = React.lazy(() => import("@/components/map/grid-view"))

const rowCount = 64
const columnCount = 64

export const MapTest = () => {
    const { height, width } = useWindowSize()

    // Calculate optimal segment sizes based on screen dimensions
    // Goal: ~9 segments visible (3x3 grid) for optimal performance
    const tileSize = 64 // Approximate tile size in pixels
    const optimalSegmentPixels = Math.min(width || 1920, height || 1080) / 2.5 // Larger segments for fewer total
    const optimalSegmentTiles = Math.max(
        16,
        Math.min(64, Math.floor(optimalSegmentPixels / tileSize)),
    )

    const { world, isLoading, progress, error } = useFetchedWorld(
        rowCount,
        columnCount,
        optimalSegmentTiles, // maxRowsPerSegment
        optimalSegmentTiles, // maxColumnsPerSegment
    )

    if (!height || !width) {
        return null
    }

    if (isLoading) {
        return <ProgressView progress={progress} />
    }

    if (world !== undefined) {
        return <Map height={height} width={width} world={world} />
    }

    return <ErrorView error={error ?? new Error("unknown error")} />
}

export default MapTest
