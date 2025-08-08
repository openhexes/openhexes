import { useTileDimensions } from "@/hooks/use-tiles"
import * as tileUtil from "@/lib/tiles"
import { create } from "@bufbuild/protobuf"
import { useWindowSize } from "@uidotdev/usehooks"
import { type Grid, type Tile as PTile, Segment_BoundsSchema } from "proto/ts/map/v1/tile_pb"
import React from "react"

import { PatternLayer } from "./pattern-layer"
import { TileView } from "./tile-view"

interface MapProps {
    grid: Grid
}

interface Position {
    x: number
    y: number
}

const HATCH_SVG = `<svg xmlns='http://www.w3.org/2000/svg' width='12' height='12'>
  <g transform='rotate(45 6 6)'><rect x='-6' y='5' width='24' height='2' fill='#27c46b'/></g></svg>`

const WATER_SVG = `
<svg xmlns='http://www.w3.org/2000/svg' width='16' height='16'>
  <path d='M0 5 Q4 1 8 5 T16 5 M0 13 Q4 9 8 13 T16 13' fill='none' stroke='#57c0ff' stroke-width='1'/>
</svg>`

export const GridView: React.FC<MapProps> = ({ grid }) => {
    const windowSize = useWindowSize()
    const { tileHeight, tileWidth, rowHeight, triangleHeight } = useTileDimensions()

    const rafRef = React.useRef<number | null>(null)
    const pending = React.useRef<{ dx: number; dy: number } | null>(null)

    React.useEffect(() => {
        // prevent "go back/forward" on overscroll
        const prev = document.body.style.overscrollBehaviorX
        document.body.style.overscrollBehaviorX = "none"
        return () => {
            document.body.style.overscrollBehaviorX = prev
        }
    }, [])

    const containerRef = React.useRef<HTMLDivElement>(null)
    const [lastPosition, setLastPosition] = React.useState<Position | null>(null)
    const [offset, setOffset] = React.useState<Position>({ x: 0, y: 0 })

    const [visibleTiles, setVisibleTiles] = React.useState<PTile[]>([])

    const mapHeight = Math.ceil((((grid.totalRows + 0.4) * tileHeight) / 2) * 1.5)
    const mapWidth = (grid.totalColumns + 1) * tileWidth

    React.useEffect(() => handlePan(0, 0), [windowSize.height, windowSize.width]) // todo

    const flushPan = () => {
        if (!pending.current) return
        const { dx, dy } = pending.current
        pending.current = null
        _applyPan(dx, dy)
        rafRef.current = null
    }

    const handlePan = (dx: number, dy: number) => {
        pending.current = {
            dx: (pending.current?.dx ?? 0) + dx,
            dy: (pending.current?.dy ?? 0) + dy,
        }
        if (rafRef.current == null) {
            rafRef.current = requestAnimationFrame(flushPan)
        }
    }

    const _applyPan = (dx: number, dy: number) => {
        const rect = containerRef.current?.getBoundingClientRect() ?? {
            height: window.innerHeight,
            width: window.innerWidth,
        }

        setOffset((prev) => {
            const next = { x: prev.x, y: prev.y }
            if (rect.width < mapWidth) {
                const maxX = mapWidth - rect.width
                next.x = Math.floor(Math.max(-maxX, Math.min(0, prev.x + dx)))
            } else {
                next.x = 0
            }

            if (rect.height < mapHeight) {
                const maxY = mapHeight - rect.height
                next.y = Math.floor(Math.max(-maxY, Math.min(0, prev.y + dy)))
            } else {
                next.y = 0
            }

            return next
        })

        const maxVisibleRows = Math.ceil((rect.height - 2 * triangleHeight) / rowHeight)
        const maxVisibleColumns = Math.ceil(rect.width / tileWidth)
        const skippedColumnCount = Math.ceil(-offset.x / tileWidth)
        const skippedRowCount = Math.ceil(-offset.y / rowHeight)
        const visibleBounds = create(Segment_BoundsSchema, {
            minRow: skippedRowCount,
            maxRow: skippedRowCount + maxVisibleRows,
            minColumn: skippedColumnCount,
            maxColumn: skippedColumnCount + maxVisibleColumns,
        })

        const rowStartingIndex =
            Math.floor((skippedRowCount / grid.totalRows) * grid.segmentRows.length) - 1

        const visibleTiles: PTile[] = []
        for (let i = rowStartingIndex; i < grid.segmentRows.length; i++) {
            const row = grid.segmentRows[i]
            if (row === undefined) {
                continue
            }

            let segmentsFound = false
            const columnStartingIndex =
                Math.floor((skippedColumnCount / grid.totalColumns) * row.segments.length) - 1

            for (let j = columnStartingIndex; j < row.segments.length; j++) {
                const segment = row.segments[j]
                if (segment === undefined) {
                    continue
                }

                if (tileUtil.boundsIntersect(segment.bounds, visibleBounds)) {
                    visibleTiles.push(
                        ...segment.tiles.filter((tile) =>
                            tileUtil.boundsInclude(tile, visibleBounds, 2),
                        ),
                    )
                    segmentsFound = true
                } else {
                    if (segmentsFound) {
                        break
                    }
                }
            }

            if (!segmentsFound && visibleTiles.length > 0) {
                break
            }
        }
        setVisibleTiles(visibleTiles)
    }

    const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        e.stopPropagation()

        if (!e.ctrlKey) {
            setLastPosition(null)
            return
        }

        const x = e.clientX
        const y = e.clientY
        if (lastPosition) {
            const dx = x - lastPosition.x
            const dy = y - lastPosition.y
            handlePan(dx, dy)
        }
        setLastPosition({ x, y })
    }

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        e.stopPropagation()
        handlePan(-e.deltaX, -e.deltaY)
    }

    const handleTouchMove = (e: React.TouchEvent<HTMLDivElement>) => {
        e.stopPropagation()

        const x = e.touches[0].clientX
        const y = e.touches[0].clientY
        if (lastPosition) {
            const dx = x - lastPosition.x
            const dy = y - lastPosition.y
            handlePan(dx, dy)
        }
        setLastPosition({ x, y })
    }

    const handleTouchEnd = () => {
        setLastPosition(null)
    }

    return (
        <div
            ref={containerRef}
            data-testid="map-container"
            className="overflow-hidden h-screen w-screen"
            onMouseMoveCapture={handleMouseMove}
            onWheelCapture={handleWheel}
            onTouchMoveCapture={handleTouchMove}
            onTouchEndCapture={handleTouchEnd}
        >
            <div
                data-testid="map"
                className="select-none cursor-pointer"
                style={{
                    transform: `translate(${offset.x}px, ${offset.y}px)`,
                }}
            >
                <PatternLayer
                    id="water"
                    tiles={visibleTiles}
                    filter={(t) => t.terrainId === "core/terrain/water"}
                    tileWidth={tileWidth}
                    tileHeight={tileHeight}
                    mapWidth={mapWidth}
                    mapHeight={mapHeight}
                    cell={24}
                    svgTile={WATER_SVG}
                    opacity={0.35}
                />
                <PatternLayer
                    id="forest"
                    tiles={visibleTiles}
                    filter={(t) =>
                        t.terrainId === "core/terrain/forest" ||
                        t.terrainId === "core/terrain/grass"
                    }
                    tileWidth={tileWidth}
                    tileHeight={tileHeight}
                    mapWidth={mapWidth}
                    mapHeight={mapHeight}
                    cell={32}
                    svgTile={HATCH_SVG}
                    opacity={0.4}
                />

                {visibleTiles.map((tile) => (
                    <TileView tile={tile} key={tileUtil.getKey(tile)} />
                ))}
            </div>
        </div>
    )
}

export default GridView
