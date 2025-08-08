import { cn } from "@/lib/utils"
import { Check, Clock, Loader2 } from "lucide-react"
import { type Progress as Proto, Stage_State } from "proto/ts/progress/v1/progress_pb"
import React from "react"

import { Progress } from "../ui/progress"

interface P {
    progress?: Proto
}

export const ProgressView: React.FC<P> = ({ progress }) => (
    <div className="flex flex-col gap-6 p-6 justify-center items-center h-screen w-screen">
        <Progress value={(progress?.percentage ?? 0) * 100} className="w-sm" />
        <div className="flex flex-col gap-2 w-sm">
            {progress?.stages?.map((s, i) => (
                <div key={i} className="flex gap-2 items-center text-sm">
                    {s.state === Stage_State.RUNNING && (
                        <Loader2 size={16} className="animate-spin text-muted-foreground" />
                    )}
                    {s.state === Stage_State.DONE && <Check size={16} className="text-green-600" />}
                    {s.state === Stage_State.WAITING && (
                        <Clock size={16} className="text-muted-foreground" />
                    )}
                    <div
                        className={cn("flex gap-1 items-center justify-between w-full", {
                            "text-muted-foreground": s.state === Stage_State.WAITING,
                        })}
                    >
                        {s.title}
                        {s.duration === undefined && s.subtitle && (
                            <div className="text-xs text-muted-foreground">{s.subtitle}</div>
                        )}
                        {s.duration && (
                            <div className="text-xs text-muted-foreground">
                                {(
                                    Number(s.duration.seconds) * 1000 +
                                    s.duration.nanos / 1_000_000
                                ).toFixed(2)}{" "}
                                ms
                            </div>
                        )}
                    </div>
                </div>
            ))}
        </div>
    </div>
)
