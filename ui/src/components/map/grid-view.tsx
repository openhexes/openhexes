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
    const [, setAccumulatedDelta] = React.useState<Position>({ x: 0, y: 0 })
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
            // Early exit if no movement
            if (dx === 0 && dy === 0) return
            
            const rect = containerRef.current?.getBoundingClientRect() ?? {
                height: window.innerHeight,
                width: window.innerWidth,
            }

            // Accumulate movement deltas
            setAccumulatedDelta((prev) => {
                const newAccumulated = {
                    x: prev.x + dx,
                    y: prev.y + dy,
                }

                const effectiveRowHeight = rowHeight * zoom
                const effectiveTileWidth = tileWidth * zoom

                // Calculate how many tiles to move based on accumulated delta
                const tilesToMoveX = Math.floor(Math.abs(newAccumulated.x) / effectiveTileWidth)
                const tilesToMoveY = Math.floor(Math.abs(newAccumulated.y) / effectiveRowHeight)

                // Early exit if no tile movement - don't trigger any state updates
                if (tilesToMoveX === 0 && tilesToMoveY === 0) {
                    return newAccumulated // Return accumulated delta but don't update offset
                }

                // Calculate actual pixel movement (discrete tile steps)
                const actualDx = tilesToMoveX * effectiveTileWidth * Math.sign(newAccumulated.x)
                const actualDy = tilesToMoveY * effectiveRowHeight * Math.sign(newAccumulated.y)

                // Apply the discrete pan movement
                setOffset((prev) => {
                    const next = { x: prev.x, y: prev.y }

                    // Account for zoom when calculating pan bounds
                    const scaledMapWidth = mapWidth * zoom
                    const scaledMapHeight = mapHeight * zoom

                    if (rect.width < scaledMapWidth) {
                        const maxX = scaledMapWidth - rect.width
                        const newX = Math.max(-maxX, Math.min(0, prev.x + actualDx))
                        next.x = newX
                    } else {
                        // Center the map if it's smaller than viewport
                        next.x = (rect.width - scaledMapWidth) / 2
                    }

                    if (rect.height < scaledMapHeight) {
                        const maxY = scaledMapHeight - rect.height
                        const newY = Math.max(-maxY, Math.min(0, prev.y + actualDy))
                        next.y = newY
                    } else {
                        // Center the map if it's smaller than viewport
                        next.y = (rect.height - scaledMapHeight) / 2
                    }

                    console.info("OFFSET CHANGED", { from: prev, to: next })
                    return next
                })

                // Reset accumulated delta, keeping remainder
                return {
                    x: newAccumulated.x - actualDx,
                    y: newAccumulated.y - actualDy,
                }
            })
        },
        [mapHeight, mapWidth, rowHeight, tileWidth, zoom],
    )

    // Separate effect for visibility calculation - only runs when offset actually changes
    React.useEffect(() => {
        const rect = containerRef.current?.getBoundingClientRect() ?? {
            height: window.innerHeight,
            width: window.innerWidth,
        }

        // Account for zoom when calculating visibility
        const effectiveRowHeight = rowHeight * zoom
        const effectiveTileWidth = tileWidth * zoom
        // Small buffer for edge cases, but not too large
        const maxVisibleRows = Math.ceil(rect.height / effectiveRowHeight) + 1
        const maxVisibleColumns = Math.ceil(rect.width / effectiveTileWidth) + 1

        // Simple tile-based calculation
        const startColumn = Math.max(0, Math.floor(-offset.x / effectiveTileWidth))
        const startRow = Math.max(0, Math.floor(-offset.y / effectiveRowHeight))

        const visibleBounds = create(Segment_BoundsSchema, {
            minRow: startRow,
            maxRow: startRow + maxVisibleRows,
            minColumn: startColumn,
            maxColumn: startColumn + maxVisibleColumns,
        })

        const visibleSegments: Segment[] = []
        const visibleTiles: PTile[] = []

        // Check all segments and track visibility
        for (const row of grid.segmentRows) {
            if (!row) continue

            for (const segment of row.segments) {
                if (!segment || !segment.bounds) continue

                const intersects = tileUtil.boundsIntersect(segment.bounds, visibleBounds)

                if (intersects) {
                    visibleSegments.push(segment)
                    visibleTiles.push(...(segment.tiles || []))
                }
            }
        }

        console.info("UPDATED SEGMENTS")
        setVisibleSegments(visibleSegments)
        setVisibleTiles(visibleTiles)
    }, [offset.x, offset.y, rowHeight, tileWidth, zoom, grid.segmentRows])

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

    // Visibility is now handled by the effect above, no need for separate effect

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

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        // Invert direction for natural trackpad scrolling
        handlePan(-e.deltaX, -e.deltaY)
    }

    const handleClick = React.useCallback(
        (e: React.MouseEvent<HTMLDivElement>) => {
            if (e.shiftKey) {
                e.preventDefault()
                e.stopPropagation()

                const newZoomState = !isZoomedOut
                setIsZoomedOut(newZoomState)
                setAccumulatedDelta({ x: 0, y: 0 }) // Reset accumulated movement

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
                onWheelCapture={handleWheel}
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
                    {/* Actual segments for rendering */}
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
