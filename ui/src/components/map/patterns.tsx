import type React from "react"

import { PatternLayer } from "./pattern-layer"
import leaves from "./patterns/leaves.svg?raw"
import plus from "./patterns/plus.svg?raw"
import mountains from "./patterns/terazzo.svg?raw"
import waves from "./patterns/waves.svg?raw"

type P = Omit<
    React.ComponentProps<typeof PatternLayer>,
    "id" | "cellW" | "cellH" | "svgTile" | "backgroundColor" | "opacity"
>

export const WavesPattern: React.FC<P> = (props) => (
    <PatternLayer
        id="waves"
        cellW={100}
        cellH={20}
        svgTile={waves}
        backgroundColor="--color-sky-800"
        opacity={0.5}
        {...props}
    />
)

export const LeavesPattern: React.FC<P> = (props) => (
    <PatternLayer
        id="leaves"
        cellW={80}
        cellH={40}
        svgTile={leaves}
        backgroundColor="--color-green-950"
        opacity={0.7}
        {...props}
    />
)

export const MountainsPattern: React.FC<P> = (props) => (
    <PatternLayer
        id="mountains"
        cellW={200}
        cellH={200}
        svgTile={mountains}
        backgroundColor="--color-stone-800"
        opacity={0.3}
        {...props}
    />
)

export const PlusPattern: React.FC<P> = (props) => (
    <PatternLayer
        id="plus"
        cellW={60}
        cellH={60}
        svgTile={plus}
        backgroundColor="--color-zinc-950"
        opacity={1}
        {...props}
    />
)
