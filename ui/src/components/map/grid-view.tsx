import { WorldContext } from "@/hooks/use-world"
import * as segmentUtil from "@/lib/segments"
import * as tileUtil from "@/lib/tiles"
import { create } from "@bufbuild/protobuf"
import {
    type Grid,
    type Tile as PTile,
    type Segment,
    Segment_BoundsSchema,
} from "proto/ts/map/v1/tile_pb"
import type { World } from "proto/ts/world/v1/world_pb"
import React from "react"

import { PatternLayer } from "./pattern-layer"
import { StatusBar } from "./status-bar"
import { TileView } from "./tile-view"

interface MapProps {
    height: number
    width: number
    world: World
    grid: Grid
}

interface Position {
    x: number
    y: number
}

export const GridView: React.FC<MapProps> = ({ height, width, world, grid }) => {
    const { tileHeight = 0, tileWidth = 0 } = world.renderingSpec || {}

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

    const [visibleSegments, setVisibleSegments] = React.useState<Segment[]>([])
    const [visibleTiles, setVisibleTiles] = React.useState<PTile[]>([])

    const rowHeight = tileHeight * 0.75
    const mapWidth = Math.ceil(grid.totalColumns * tileWidth + tileWidth / 2) // extra half for shifted rows
    const mapHeight = Math.ceil((grid.totalRows - 1) * rowHeight + tileHeight) // = tileHeight*(0.75*rows + 0.25)

    const _applyPan = React.useCallback(
        (dx: number, dy: number) => {
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

            // const maxVisibleRows = Math.ceil((rect.height - 2 * triangleHeight) / rowHeight)
            const maxVisibleRows = Math.ceil(rect.height / rowHeight)
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

            const visibleSegments: Segment[] = []
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
                        visibleSegments.push(segment)
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
            setVisibleSegments(visibleSegments)
            setVisibleTiles(visibleTiles)
        },
        [
            grid.segmentRows,
            grid.totalColumns,
            grid.totalRows,
            mapHeight,
            mapWidth,
            offset.x,
            offset.y,
            rowHeight,
            tileWidth,
        ],
    )

    const flushPan = React.useCallback(() => {
        if (!pending.current) return
        const { dx, dy } = pending.current
        pending.current = null
        _applyPan(dx, dy)
        rafRef.current = null
    }, [_applyPan])

    const handlePan = React.useCallback(
        (dx: number, dy: number) => {
            pending.current = {
                dx: (pending.current?.dx ?? 0) + dx,
                dy: (pending.current?.dy ?? 0) + dy,
            }
            if (rafRef.current == null) {
                rafRef.current = requestAnimationFrame(flushPan)
            }
        },
        [flushPan],
    )

    React.useEffect(() => handlePan(0, 0), [height, width, handlePan])

    const [selectedTile, setSelectedTile] = React.useState<PTile | undefined>(undefined)

    const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        e.stopPropagation()

        if (!e.metaKey) {
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

    const handleTileSelect = (tile?: PTile) => {
        if (tile?.key === selectedTile?.key) {
            setSelectedTile(undefined)
        } else {
            setSelectedTile(tile)
        }
    }

    return (
        <WorldContext value={{ world: world, selectedTile, selectTile: handleTileSelect }}>
            <div
                ref={containerRef}
                data-testid="map-container"
                className="overflow-hidden"
                style={{ height, width }}
                onMouseMoveCapture={handleMouseMove}
                onTouchMoveCapture={handleTouchMove}
                onTouchEndCapture={handleTouchEnd}
                onMouseLeave={() => handleTileSelect()}
            >
                <div
                    data-testid="map"
                    className="relative select-none cursor-pointer"
                    style={{
                        width: mapWidth,
                        height: mapHeight,
                        transform: `translate(${offset.x}px, ${offset.y}px)`,
                        willChange: "transform",
                    }}
                >
                    {visibleSegments.map((segment) => (
                        <PatternLayer
                            key={segmentUtil.getKey(segment)}
                            segment={segment}
                            tileHeight={tileHeight}
                            tileWidth={tileWidth}
                        />
                    ))}

                    {/* interactive tiles on top */}
                    {visibleTiles.map((tile) => (
                        <TileView
                            tile={tile}
                            key={tile.key}
                            height={tileHeight}
                            width={tileWidth}
                        />
                    ))}
                </div>

                <StatusBar />
            </div>
        </WorldContext>
    )
}

export default GridView
