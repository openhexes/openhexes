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
// REMOVED: import { TileView } from "./tile-view" - not needed, no individual interactive tiles

import "./tile.css"

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
    const [smoothPan, setSmoothPan] = React.useState<boolean>(false)

    const panRef = React.useRef<number | null>(null)
    const panPending = React.useRef<{ dx: number; dy: number } | null>(null)

    const containerRef = React.useRef<HTMLDivElement>(null)
    const lastPositionRef = React.useRef<Position | null>(null)
    const [offset, setOffset] = React.useState<Position>({ x: 0, y: 0 })
    const accumulatedDeltaRef = React.useRef<Position>({ x: 0, y: 0 })
    const [isZoomedOut, setIsZoomedOut] = React.useState<boolean>(false)

    const visibleSegmentsRef = React.useRef<Segment[]>([])
    const [, forceRender] = React.useState({})
    // REMOVED: const [visibleTiles, setVisibleTiles] = React.useState<PTile[]>([]) - not needed without TileView
    
    // Custom function to update visible segments only when they actually change
    const updateVisibleSegments = (newSegments: Segment[]) => {
        const current = visibleSegmentsRef.current
        
        // Fast path: if lengths are different, definitely changed
        if (current.length !== newSegments.length) {
            visibleSegmentsRef.current = newSegments
            forceRender({}) // Trigger re-render
            return
        }
        
        // Check if segments actually changed (by reference comparison)
        let hasChanged = false
        for (let i = 0; i < current.length; i++) {
            if (current[i] !== newSegments[i]) {
                hasChanged = true
                break
            }
        }
        
        if (hasChanged) {
            visibleSegmentsRef.current = newSegments
            forceRender({}) // Only re-render if segments actually changed
        }
    }
    // REMOVED: const [visibleTiles, setVisibleTiles] = React.useState<PTile[]>([])

    const { tileHeight = 0, tileWidth = 0 } = world.renderingSpec || {}
    const grid = world.layers[selectedDepth]
    const rowHeight = tileHeight * 0.75
    const mapWidth = Math.ceil(grid.totalColumns * tileWidth + tileWidth / 2)
    const mapHeight = Math.ceil((grid.totalRows - 1) * rowHeight + tileHeight)

    // Calculate zoom level for fit-to-screen
    const rect = containerRef.current?.getBoundingClientRect()
    const fitZoom = rect ? Math.min(rect.width / mapWidth, rect.height / mapHeight) : 0.5
    const zoom = isZoomedOut ? fitZoom : 1

    const _applyPan = (dx: number, dy: number) => {
        // Early exit if no movement
        if (dx === 0 && dy === 0) return
        
        const rect = containerRef.current?.getBoundingClientRect() ?? {
            height: window.innerHeight,
            width: window.innerWidth,
        }

        if (smoothPan) {
            // Smooth panning - apply movement directly
            setOffset((prev) => {
                const next = { x: prev.x + dx, y: prev.y + dy }
                const scaledMapWidth = mapWidth * zoom
                const scaledMapHeight = mapHeight * zoom

                if (rect.width < scaledMapWidth) {
                    const maxX = scaledMapWidth - rect.width
                    next.x = Math.max(-maxX, Math.min(0, next.x))
                } else {
                    next.x = (rect.width - scaledMapWidth) / 2
                }

                if (rect.height < scaledMapHeight) {
                    const maxY = scaledMapHeight - rect.height
                    next.y = Math.max(-maxY, Math.min(0, next.y))
                } else {
                    next.y = (rect.height - scaledMapHeight) / 2
                }

                return next
            })
        } else {
            // Discrete panning - accumulate deltas and move in tile steps
            const newAccumulated = {
                x: accumulatedDeltaRef.current.x + dx,
                y: accumulatedDeltaRef.current.y + dy,
            }

            const effectiveRowHeight = rowHeight * zoom
            const effectiveTileWidth = tileWidth * zoom

            const tilesToMoveX = Math.floor(Math.abs(newAccumulated.x) / effectiveTileWidth)
            const tilesToMoveY = Math.floor(Math.abs(newAccumulated.y) / effectiveRowHeight)

            if (tilesToMoveX === 0 && tilesToMoveY === 0) {
                accumulatedDeltaRef.current = newAccumulated
                return
            }

            const actualDx = tilesToMoveX * effectiveTileWidth * Math.sign(newAccumulated.x)
            const actualDy = tilesToMoveY * effectiveRowHeight * Math.sign(newAccumulated.y)

            setOffset((prev) => {
                const next = { x: prev.x, y: prev.y }
                const scaledMapWidth = mapWidth * zoom
                const scaledMapHeight = mapHeight * zoom

                if (rect.width < scaledMapWidth) {
                    const maxX = scaledMapWidth - rect.width
                    const newX = Math.max(-maxX, Math.min(0, prev.x + actualDx))
                    next.x = newX
                } else {
                    next.x = (rect.width - scaledMapWidth) / 2
                }

                if (rect.height < scaledMapHeight) {
                    const maxY = scaledMapHeight - rect.height
                    const newY = Math.max(-maxY, Math.min(0, prev.y + actualDy))
                    next.y = newY
                } else {
                    next.y = (rect.height - scaledMapHeight) / 2
                }

                return next
            })

            accumulatedDeltaRef.current = {
                x: newAccumulated.x - actualDx,
                y: newAccumulated.y - actualDy,
            }
        }
    }

    // Simple visibility calculation - only runs when offset changes
    React.useEffect(() => {
        const rect = containerRef.current?.getBoundingClientRect() ?? {
            height: window.innerHeight,
            width: window.innerWidth,
        }

        const effectiveRowHeight = rowHeight * zoom
        const effectiveTileWidth = tileWidth * zoom
        const maxVisibleRows = Math.ceil(rect.height / effectiveRowHeight) + 1
        const maxVisibleColumns = Math.ceil(rect.width / effectiveTileWidth) + 1

        const startColumn = Math.max(0, Math.floor(-offset.x / effectiveTileWidth))
        const startRow = Math.max(0, Math.floor(-offset.y / effectiveRowHeight))

        const visibleBounds = create(Segment_BoundsSchema, {
            minRow: startRow,
            maxRow: startRow + maxVisibleRows,
            minColumn: startColumn,
            maxColumn: startColumn + maxVisibleColumns,
        })

        const newVisibleSegments: Segment[] = []

        // Get actual segment dimensions from the world data
        // Don't hardcode segment sizes - use what's actually in the data
        const firstSegment = grid.segmentRows[0]?.segments[0]
        const actualMaxRows = firstSegment?.bounds ? (firstSegment.bounds.maxRow - firstSegment.bounds.minRow) : 16
        const actualMaxCols = firstSegment?.bounds ? (firstSegment.bounds.maxColumn - firstSegment.bounds.minColumn) : 20
        
        // Optimize: only check segment rows that could potentially be visible
        const segmentRowStart = Math.max(0, Math.floor(startRow / actualMaxRows))
        const segmentRowEnd = Math.min(grid.segmentRows.length - 1, Math.ceil((startRow + maxVisibleRows) / actualMaxRows))

        for (let segmentRowIndex = segmentRowStart; segmentRowIndex <= segmentRowEnd; segmentRowIndex++) {
            const row = grid.segmentRows[segmentRowIndex]
            if (!row) continue
            
            // Similarly, optimize segment columns within each row
            const segmentColStart = Math.max(0, Math.floor(startColumn / actualMaxCols))
            const segmentColEnd = Math.min(row.segments.length - 1, Math.ceil((startColumn + maxVisibleColumns) / actualMaxCols))
            
            for (let segmentColIndex = segmentColStart; segmentColIndex <= segmentColEnd; segmentColIndex++) {
                const segment = row.segments[segmentColIndex]
                if (!segment?.bounds) continue
                if (tileUtil.boundsIntersect(segment.bounds, visibleBounds)) {
                    newVisibleSegments.push(segment)
                }
            }
        }

        updateVisibleSegments(newVisibleSegments)
    }, [offset.x, offset.y, rowHeight, tileWidth, zoom, grid.segmentRows, updateVisibleSegments])

    const flushPan = () => {
        if (!panPending.current) return
        const { dx, dy } = panPending.current
        panPending.current = null
        _applyPan(dx, dy)
        panRef.current = null
    }

    const handlePan = (dx: number, dy: number) => {
        panPending.current = {
            dx: (panPending.current?.dx ?? 0) + dx,
            dy: (panPending.current?.dy ?? 0) + dy,
        }
        if (panRef.current == null) {
            panRef.current = requestAnimationFrame(flushPan)
        }
    }

    // Visibility is now handled by the effect above, no need for separate effect

    React.useEffect(() => handlePan(0, 0), [height, width])

    const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
        e.preventDefault()
        e.stopPropagation()

        // Handle panning if meta key is pressed
        if (e.metaKey) {
            // Clear hover during panning
            if (!lastPositionRef.current) {
                setSelectedTile(undefined)
            }
            
            const x = e.clientX
            const y = e.clientY
            if (lastPositionRef.current) {
                const dx = x - lastPositionRef.current.x
                const dy = y - lastPositionRef.current.y
                handlePan(dx, dy)
            }
            lastPositionRef.current = { x, y }
            return
        }

        // Not panning - reset pan state
        lastPositionRef.current = null

        // Hover detection for tile selection
        if (!isZoomedOut) {
            const rect = containerRef.current?.getBoundingClientRect()
            if (!rect) return

            const relativeX = e.clientX - rect.left - offset.x
            const relativeY = e.clientY - rect.top - offset.y

            const scaledX = relativeX / zoom
            const scaledY = relativeY / zoom

            const row = Math.floor(scaledY / rowHeight)
            const col = Math.floor((scaledX - (row % 2 === 0 ? 0 : tileWidth / 2)) / tileWidth)

            if (row >= 0 && row < grid.totalRows && col >= 0 && col < grid.totalColumns) {
                // Find tile in visible segments
                let foundTile = undefined
                for (const segment of visibleSegmentsRef.current) {
                    if (!segment.tiles) continue
                    const tile = segment.tiles.find(tile => {
                        if (!tile.coordinate) return false
                        return tile.coordinate.row === row && tile.coordinate.column === col
                    })
                    if (tile) {
                        foundTile = tile
                        break
                    }
                }
                setSelectedTile(foundTile)
            } else {
                setSelectedTile(undefined)
            }
        }
    }

    const handleTouchMove = (e: React.TouchEvent<HTMLDivElement>) => {
        e.stopPropagation()

        const x = e.touches[0].clientX
        const y = e.touches[0].clientY
        if (lastPositionRef.current) {
            const dx = x - lastPositionRef.current.x
            const dy = y - lastPositionRef.current.y
            handlePan(dx, dy)
        }
        lastPositionRef.current = { x, y }
    }

    const handleTouchEnd = () => {
        lastPositionRef.current = null
    }

    const handleWheel = (e: React.WheelEvent<HTMLDivElement>) => {
        e.preventDefault() // Prevent page bounce
        e.stopPropagation() // Stop event bubbling
        
        // Invert direction for natural trackpad scrolling
        handlePan(-e.deltaX, -e.deltaY)
    }

    const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
        if (e.shiftKey) {
            e.preventDefault()
            e.stopPropagation()

            const newZoomState = !isZoomedOut
            setIsZoomedOut(newZoomState)
            accumulatedDeltaRef.current = { x: 0, y: 0 } // Reset accumulated movement

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
    }

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
                smoothPan,
                setSmoothPan,
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
                    {visibleSegmentsRef.current.map((segment) => (
                        <PatternLayer
                            key={segmentUtil.getKey(segment)}
                            segment={segment}
                            tileHeight={tileHeight}
                            tileWidth={tileWidth}
                        />
                    ))}

                    {/* Tile highlight overlay */}
                    {selectedTile && (() => {
                        const { row, column } = tileUtil.getCoordinates(selectedTile)
                        const even = row % 2 === 0
                        const left = column * tileWidth + (even ? 0 : tileWidth / 2)
                        const top = row * rowHeight
                        
                        // Create hex border using SVG - the only way to get proper hex border
                        const hexPoints = `
                            ${tileWidth / 2},2 
                            ${tileWidth - 2},${tileHeight / 4} 
                            ${tileWidth - 2},${(3 * tileHeight) / 4} 
                            ${tileWidth / 2},${tileHeight - 2} 
                            2,${(3 * tileHeight) / 4} 
                            2,${tileHeight / 4}
                        `
                        
                        return (
                            <div
                                className="absolute pointer-events-none"
                                style={{
                                    left,
                                    top,
                                    width: tileWidth,
                                    height: tileHeight,
                                    zIndex: 999,
                                }}
                            >
                                <svg width={tileWidth} height={tileHeight}>
                                    <polygon
                                        points={hexPoints}
                                        fill="rgba(253, 224, 71, 0.3)"
                                        stroke="#eab308"
                                        strokeWidth="2"
                                    />
                                </svg>
                            </div>
                        )
                    })()}
                </div>

                <StatusBar />
            </div>
        </WorldContext>
    )
}

export default GridView
