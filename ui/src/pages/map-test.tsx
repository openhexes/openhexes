import { Progress } from "@/components/ui/progress"
import { ErrorView } from "@/components/utils/error"
import { useTileGrid } from "@/hooks/use-tiles"
import { cn } from "@/lib/utils"
import { Check, Clock, Loader2 } from "lucide-react"
import { Stage_State } from "proto/ts/progress/v1/progress_pb"
import React from "react"

const Map = React.lazy(() => import("@/components/map/grid-view"))

const rowCount = 256
const columnCount = 256

export const MapTest = () => {
    const { grid, isLoading, progress, error } = useTileGrid(rowCount, columnCount, 30, 30)

    if (isLoading) {
        return (
            <div className="flex flex-col gap-6 p-6 justify-center items-center h-screen w-screen">
                <Progress value={(progress?.percentage ?? 0) * 100} className="w-sm" />
                <div className="flex flex-col gap-2 w-sm">
                    {progress?.stages?.map((s, i) => (
                        <div key={i} className="flex gap-2 items-center text-sm">
                            {s.state === Stage_State.RUNNING && (
                                <Loader2 size={16} className="animate-spin text-muted-foreground" />
                            )}
                            {s.state === Stage_State.DONE && (
                                <Check size={16} className="text-green-600" />
                            )}
                            {s.state === Stage_State.WAITING && (
                                <Clock size={16} className="text-muted-foreground" />
                            )}
                            <div
                                className={cn("flex gap-1 items-center justify-between w-full", {
                                    "text-muted-foreground": s.state === Stage_State.WAITING,
                                })}
                            >
                                {s.title}
                                {s.subtitle && (
                                    <div className="text-xs text-muted-foreground">
                                        {s.subtitle}
                                    </div>
                                )}
                                {/* {s.duration && (
                                    <div className="text-xs text-muted-foreground">
                                        {s.duration.seconds}s
                                    </div>
                                )} */}
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
