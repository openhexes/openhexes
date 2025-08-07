import React from "react"
import { createRoot } from "react-dom/client"
import { Route } from "react-router"

import type { TileProps } from "./components/map/tile"
import "./index.css"

const Layout = React.lazy(() => import("@/components/utils/layout"))
const NotFound = React.lazy(() => import("@/components/utils/not-found"))
const Map = React.lazy(() => import("@/components/map/map"))

const tiles: TileProps[] = []
for (let row = 0; row < 31; row++) {
    for (let column = 0; column < 35; column++) {
        tiles.push({ row, column, visible: false })
    }
}

let maxRow = 0
let maxColumn = 0
for (const tile of tiles) {
    maxRow = Math.max(tile.row, maxRow)
    maxColumn = Math.max(tile.column, maxColumn)
}

createRoot(document.getElementById("root")!).render(
    <Layout>
        <Route path="/" element={<Map tiles={tiles} maxColumn={maxColumn} maxRow={maxRow} />} />
        <Route path="*" element={<NotFound />} />
    </Layout>,
)
