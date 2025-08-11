import { useNoOverscroll } from "@/hooks/use-no-overscroll"
import { WorldContext } from "@/hooks/use-world"
import * as segmentUtil from "@/lib/segments"
import * as tileUtil from "@/lib/tiles"
import { create } from "@bufbuild/protobuf"
import { type Tile as PTile, type Segment, Segment_BoundsSchema } from "proto/ts/map/v1/tile_pb"
import type { World } from "proto/ts/world/v1/world_pb"
import React from "react"

import { PatternLayer } from "./pattern-layer"
import { StatusBar } from "./status-bar"
import { TileView } from "./tile-view"

interface MapProps {
    height: number
    width: number
    world: World
}

interface Position {
    x: number
    y: number
}

export const GridView: React.FC<MapProps> = ({ height, width, world }) => {
    useNoOverscroll()

    const [selectedTile, setSelectedTile] = React.useState<PTile | undefined>(undefined)
    const [selectedDepth, selectDepth] = React.useState<number>(0)

    const panRef = React.useRef<number | null>(null)
    const panPending = React.useRef<{ dx: number; dy: number } | null>(null)

    const containerRef = React.useRef<HTMLDivElement>(null)
    const [lastPosition, setLastPosition] = React.useState<Position | null>(null)
    const [offset, setOffset] = React.useState<Position>({ x: 0, y: 0 })
    const [isZoomedOut, setIsZoomedOut] = React.useState<boolean>(false)

    const [visibleSegments, setVisibleSegments] = React.useState<Segment[]>([])
    const [visibleTiles, setVisibleTiles] = React.useState<PTile[]>([])

    const { tileHeight = 0, tileWidth = 0 } = world.renderingSpec || {}
    const grid = world.layers[selectedDepth]
    const rowHeight = tileHeight * 0.75
    const mapWidth = Math.ceil(grid.totalColumns * tileWidth + tileWidth / 2) // extra half for shifted rows
    const mapHeight = Math.ceil((grid.totalRows - 1) * rowHeight + tileHeight) // = tileHeight*(0.75*rows + 0.25)

    // Calculate zoom level for fit-to-screen
    const rect = containerRef.current?.getBoundingClientRect()
    const fitZoom = rect ? Math.min(rect.width / mapWidth, rect.height / mapHeight) : 0.5
    const zoom = isZoomedOut ? fitZoom : 1

    const _applyPan = React.useCallback(
        (dx: number, dy: number) => {
            const rect = containerRef.current?.getBoundingClientRect() ?? {
                height: window.innerHeight,
                width: window.innerWidth,
            }

            setOffset((prev) => {
                const next = { x: prev.x, y: prev.y }

                // Account for zoom when calculating pan bounds
                const scaledMapWidth = mapWidth * zoom
                const scaledMapHeight = mapHeight * zoom

                if (rect.width < scaledMapWidth) {
                    const maxX = scaledMapWidth - rect.width
                    next.x = Math.floor(Math.max(-maxX, Math.min(0, prev.x + dx)))
                } else {
                    // Center the map if it's smaller than viewport
                    next.x = (rect.width - scaledMapWidth) / 2
                }

                if (rect.height < scaledMapHeight) {
                    const maxY = scaledMapHeight - rect.height
                    next.y = Math.floor(Math.max(-maxY, Math.min(0, prev.y + dy)))
                } else {
                    // Center the map if it's smaller than viewport
                    next.y = (rect.height - scaledMapHeight) / 2
                }

                return next
            })

            // Account for zoom when calculating visibility (more tiles visible when zoomed out)
            const effectiveRowHeight = rowHeight * zoom
            const effectiveTileWidth = tileWidth * zoom
            const maxVisibleRows = Math.ceil(rect.height / effectiveRowHeight)
            const maxVisibleColumns = Math.ceil(rect.width / effectiveTileWidth)
            const skippedColumnCount = Math.ceil(-offset.x / effectiveTileWidth)
            const skippedRowCount = Math.ceil(-offset.y / effectiveRowHeight)
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
            zoom,
        ],
    )

    const flushPan = React.useCallback(() => {
        if (!panPending.current) return
        const { dx, dy } = panPending.current
        panPending.current = null
        _applyPan(dx, dy)
        panRef.current = null
    }, [_applyPan])

    const handlePan = React.useCallback(
        (dx: number, dy: number) => {
            panPending.current = {
                dx: (panPending.current?.dx ?? 0) + dx,
                dy: (panPending.current?.dy ?? 0) + dy,
            }
            if (panRef.current == null) {
                panRef.current = requestAnimationFrame(flushPan)
            }
        },
        [flushPan],
    )

    React.useEffect(() => handlePan(0, 0), [height, width, handlePan])

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

    const handleClick = React.useCallback(
        (e: React.MouseEvent<HTMLDivElement>) => {
            if (e.shiftKey) {
                e.preventDefault()
                e.stopPropagation()

                const newZoomState = !isZoomedOut
                setIsZoomedOut(newZoomState)

                // When zooming out to fit, center the map
                if (newZoomState) {
                    const rect = containerRef.current?.getBoundingClientRect()
                    if (rect) {
                        const fitZoom = Math.min(rect.width / mapWidth, rect.height / mapHeight)
                        const scaledMapWidth = mapWidth * fitZoom
                        const scaledMapHeight = mapHeight * fitZoom

                        setOffset({
                            x: (rect.width - scaledMapWidth) / 2,
                            y: (rect.height - scaledMapHeight) / 2,
                        })
                    }
                } else {
                    // Reset to normal position when zooming back in
                    setOffset({ x: 0, y: 0 })
                }
            }
        },
        [isZoomedOut, mapWidth, mapHeight],
    )

    const handleTileSelect = (tile?: PTile) => {
        if (tile?.key === selectedTile?.key) {
            setSelectedTile(undefined)
        } else {
            setSelectedTile(tile)
        }
    }

    return (
        <WorldContext
            value={{
                world: world,
                selectedDepth,
                selectDepth,
                selectedTile,
                selectTile: handleTileSelect,
            }}
        >
            <div
                ref={containerRef}
                data-testid="map-container"
                className="overflow-hidden"
                style={{ height, width }}
                onMouseMoveCapture={handleMouseMove}
                onTouchMoveCapture={handleTouchMove}
                onTouchEndCapture={handleTouchEnd}
                onClickCapture={handleClick}
                onMouseLeave={() => handleTileSelect()}
            >
                <div
                    data-testid="map"
                    className="relative select-none cursor-pointer"
                    style={{
                        width: mapWidth,
                        height: mapHeight,
                        transform: `translate(${offset.x}px, ${offset.y}px) scale(${zoom})`,
                        transformOrigin: "0 0",
                        willChange: "transform",
                    }}
                >
                    {visibleSegments.map((segment) => (
                        <PatternLayer
                            key={segmentUtil.getKey(segment)}
                            segment={segment}
                            tileHeight={tileHeight}
                            tileWidth={tileWidth}
                            isZoomedOut={isZoomedOut}
                        />
                    ))}

                    {/* interactive tiles on top - only render when not zoomed out */}
                    {!isZoomedOut &&
                        visibleTiles.map((tile) => (
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
