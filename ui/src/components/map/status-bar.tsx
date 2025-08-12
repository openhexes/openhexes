import { useWorld } from "@/hooks/use-world"
import { cn } from "@/lib/utils"
import { Footprints, LocateFixed } from "lucide-react"
import React from "react"

import { LayerSelector } from "./layer-selector"

export const StatusBar: React.FC = () => {
    const { selectedTile, smoothPan, setSmoothPan, useDetailedSvg, setUseDetailedSvg } = useWorld()

    const pillClassName = cn(
        "bg-black text-white rounded-sm",
        "flex items-center gap-1",
        "p-2 py-1 text-xs opacity-80",
    )

    return (
        <div className="absolute left-0 bottom-0 w-screen p-1">
            <div className="w-full flex items-center gap-1">
                <div className={cn(pillClassName, "p-0")}>
                    <LayerSelector />
                </div>
                <button
                    className={cn(pillClassName, "cursor-pointer hover:opacity-100")}
                    onClick={() => setSmoothPan?.(!smoothPan)}
                >
                    {smoothPan ? "Smooth" : "Discrete"} Pan
                </button>
                <button
                    className={cn(pillClassName, "cursor-pointer hover:opacity-100")}
                    onClick={() => setUseDetailedSvg?.(!useDetailedSvg)}
                >
                    {useDetailedSvg ? "Detailed" : "Lightweight"} SVG
                </button>
                {selectedTile !== undefined && (
                    <>
                        <div className={cn(pillClassName, "font-mono")}>
                            <LocateFixed size={12} />
                            {selectedTile.key}
                        </div>
                        <div className={cn(pillClassName, "min-w-[110px]")}>
                            <Footprints size={12} />
                            <span className="capitalize">{selectedTile.terrainId}</span>
                        </div>
                    </>
                )}
            </div>
        </div>
    )
}
