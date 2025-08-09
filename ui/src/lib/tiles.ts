import { create } from "@bufbuild/protobuf"
import { EdgeDirection } from "proto/ts/map/v1/compass_pb"
import { type Segment_Bounds, Segment_BoundsSchema, type Tile } from "proto/ts/map/v1/tile_pb"

const emptyBounds = create(Segment_BoundsSchema)

export const boundsIntersect = (
    a: Segment_Bounds = emptyBounds,
    b: Segment_Bounds = emptyBounds,
): boolean => {
    if (a.maxRow < b.minRow || a.minRow > b.maxRow) {
        return false
    }
    if (a.maxColumn < b.minColumn || a.minRow > b.maxRow) {
        return false
    }
    return true
}

export const getCoordinates = (p: Tile) => {
    const row = p.coordinate?.row ?? 0
    const column = p.coordinate?.column ?? 0
    const depth = p.coordinate?.depth ?? 0
    return { row, column, depth }
}

export const getKey = (p: Tile) => {
    const { row, column, depth } = getCoordinates(p)
    return `${row},${column},${depth}`
}

export const boundsInclude = (
    tile: Tile,
    policy: Segment_Bounds = emptyBounds,
    extendBoundsBy = 0,
): boolean => {
    const { row, column } = getCoordinates(tile)

    return (
        row >= policy.minRow - extendBoundsBy &&
        row < policy.maxRow + extendBoundsBy &&
        column >= policy.minColumn - extendBoundsBy &&
        column < policy.maxColumn + extendBoundsBy
    )
}

// todo: cleanup

type Pt = { x: number; y: number }

export const hexVerts = (w: number, h: number): Pt[] => {
    const v = h / 4 // shoulder inset
    return [
        { x: w / 2, y: 0 }, // 0 top
        { x: w, y: v }, // 1 upper-right
        { x: w, y: 3 * v }, // 2 lower-right
        { x: w / 2, y: h }, // 3 bottom
        { x: 0, y: 3 * v }, // 4 lower-left
        { x: 0, y: v }, // 5 upper-left
    ]
}

export const insetVerts = (outer: Pt[], k: number): Pt[] => {
    // inner = center + (outer-center) * k
    const cx = outer.reduce((s, p) => s + p.x, 0) / 6
    const cy = outer.reduce((s, p) => s + p.y, 0) / 6
    return outer.map((p) => ({ x: cx + (p.x - cx) * k, y: cy + (p.y - cy) * k }))
}

export const polyD = (a: Pt, b: Pt, bi: Pt, ai: Pt, eps = 0): string => {
    // optional tiny outward overlap (eps) to guarantee no hairline gaps
    if (eps !== 0) {
        const n = (p: Pt, q: Pt) => {
            const dx = q.x - p.x,
                dy = q.y - p.y,
                L = Math.hypot(dx, dy) || 1
            // outward normal (approx): rotate left
            return { x: -dy / L, y: dx / L }
        }
        const na = n(a, b),
            nb = n(a, b)
        a = { x: a.x + na.x * eps, y: a.y + na.y * eps }
        b = { x: b.x + nb.x * eps, y: b.y + nb.y * eps }
    }
    return `M${a.x},${a.y} L${b.x},${b.y} L${bi.x},${bi.y} L${ai.x},${ai.y} Z`
}

type EdgePaint = {
    fill: string
    under?: { fill: string; grow?: number } // grow expands by epsilon outward
}

export const edgePaint = (self: string, neigh: string): EdgePaint => {
    // tweak palette to match yours
    const water = "var(--color-sky-900)"
    const shore = "var(--color-sky-900)"
    const dark = "rgba(0,0,0,.35)"
    const rock = "#6b5f56"
    const grass = "#0ea37e"
    const sand = "#d9c97a"
    const snow = "#ffffff"
    const mud = "#6b8f5c"
    const has = (id: string, key: string) => id?.includes(key)

    if (has(self, "water") && !has(neigh, "water")) {
        return { fill: shore, under: { fill: dark, grow: 0.6 } }
    }
    if (!has(self, "water") && has(neigh, "water")) {
        return { fill: shore, under: { fill: water, grow: 0.4 } }
    }
    if (has(self, "desert") || has(neigh, "desert"))
        return { fill: sand, under: { fill: dark, grow: 0.4 } }
    if (has(self, "snow") || has(neigh, "snow"))
        return { fill: snow, under: { fill: dark, grow: 0.4 } }
    if (has(self, "grass") || has(neigh, "grass"))
        return { fill: grass, under: { fill: dark, grow: 0.4 } }
    if (has(self, "marsh") || has(neigh, "marsh") || has(self, "mud") || has(neigh, "mud"))
        return { fill: mud, under: { fill: dark, grow: 0.4 } }
    return { fill: rock, under: { fill: dark, grow: 0.4 } }
}

export const SEG_BY_DIR: Record<EdgeDirection, [number, number]> = {
    [EdgeDirection.UNSPECIFIED]: [0, 0],
    [EdgeDirection.NE]: [0, 1],
    [EdgeDirection.E]: [1, 2],
    [EdgeDirection.SE]: [2, 3],
    [EdgeDirection.SW]: [3, 4],
    [EdgeDirection.W]: [4, 5],
    [EdgeDirection.NW]: [5, 0],
}

export const shorten = (a: Pt, b: Pt, px = 2): [Pt, Pt] => {
    const dx = b.x - a.x,
        dy = b.y - a.y
    const len = Math.hypot(dx, dy) || 1
    const ux = dx / len,
        uy = dy / len
    return [
        { x: a.x + ux * px, y: a.y + uy * px },
        { x: b.x - ux * px, y: b.y - uy * px },
    ]
}

// Pick a style for a boundary (self vs neighbor terrain)
export const edgeStyle = (self: string, neigh: string) => {
    // tweak colors later; these read well on your palette
    const water = "#1f5b72"
    const shore = "#7fb8cf"
    const dark = "#101010"
    const grass = "#0ea37e"
    const sand = "#d9c97a"
    const snow = "#ffffff"
    const rock = "#6b5f56"
    const mud = "#6b8f5c"

    const is = (id: string, key: string) => id?.includes(key)

    // examples
    if (is(self, "water") && !is(neigh, "water")) {
        // outer shore + inner dark water separation
        return { width: 3, color: shore, shadow: { width: 5, color: water } }
    }
    if (!is(self, "water") && is(neigh, "water")) {
        return { width: 3, color: shore, shadow: { width: 5, color: dark } }
    }
    if (is(self, "desert") || is(neigh, "desert")) {
        return { width: 2, color: sand, shadow: { width: 4, color: dark } }
    }
    if (is(self, "snow") || is(neigh, "snow")) {
        return { width: 2, color: snow, shadow: { width: 4, color: dark } }
    }
    if (is(self, "grass") || is(neigh, "grass")) {
        return { width: 2, color: grass, shadow: { width: 4, color: dark } }
    }
    if (is(self, "waste") || is(neigh, "waste") || is(self, "rock") || is(neigh, "rock")) {
        return { width: 2, color: rock, shadow: { width: 4, color: dark } }
    }
    if (is(self, "marsh") || is(neigh, "marsh") || is(self, "mud") || is(neigh, "mud")) {
        return { width: 2, color: mud, shadow: { width: 4, color: dark } }
    }
    return { width: 2, color: dark } as const
}
