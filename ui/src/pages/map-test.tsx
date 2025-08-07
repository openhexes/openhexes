import { Progress } from "@/components/ui/progress"
import { ErrorView } from "@/components/utils/error"
import { useTileGrid } from "@/hooks/use-tiles"
import { cn } from "@/lib/utils"
import { create } from "@bufbuild/protobuf"
import { Check, Clock, Loader2 } from "lucide-react"
import { type Tile, TileSchema } from "proto/ts/map/v1/tile_pb"
import React from "react"

const Map = React.lazy(() => import("@/components/map/map"))

const rowCount = 11
const columnCount = 15

export const MapTest = () => {
    const tiles: Tile[] = []
    for (let row = 0; row < rowCount; row++) {
        for (let column = 0; column < columnCount; column++) {
            tiles.push(
                create(TileSchema, {
                    coordinate: { row, column },
                }),
            )
        }
    }

    const { grid, isLoading, progress, progressSegments, error } = useTileGrid(
        tiles,
        rowCount,
        columnCount,
        30,
        30,
    )

    if (isLoading) {
        return (
            <div className="flex flex-col gap-6 p-6 justify-center items-center h-screen w-screen">
                <Progress value={progress * 100} className="w-sm" />
                <div className="flex flex-col gap-2 w-sm">
                    {progressSegments.map((s, i) => (
                        <div key={i} className="flex gap-2 items-center text-sm">
                            {s.state === "running" && (
                                <Loader2 size={16} className="animate-spin text-muted-foreground" />
                            )}
                            {s.state === "done" && <Check size={16} className="text-green-600" />}
                            {s.state === "waiting" && (
                                <Clock size={16} className="text-muted-foreground" />
                            )}
                            <div
                                className={cn("flex gap-1 items-center justify-between w-full", {
                                    "text-muted-foreground": s.state === "waiting",
                                })}
                            >
                                {s.title}
                                {s.subtitle && !s.durationMs && (
                                    <div className="text-xs text-muted-foreground">
                                        {s.subtitle}
                                    </div>
                                )}
                                {s.durationMs && (
                                    <div className="text-xs text-muted-foreground">
                                        {s.durationMs.toFixed(0)}ms
                                    </div>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        )
    }

    if (grid !== undefined) {
        return <Map grid={grid} />
    }

    return <ErrorView error={error ?? new Error("unknown error")} />
}

export default MapTest
