import { useTiles } from "@/hooks/use-tiles"
import React from "react"

import { Tile, type TileProps } from "./tile"

interface MapProps {
    tiles: TileProps[]
    maxRow: number
    maxColumn: number
}

interface Position {
    x: number
    y: number
}

interface VisibilityState {
    minRow: number
    maxRow: number
    minColumn: number
    maxColumn: number
}

const overshot = {
    x: 0,
    y: 3,
}

export const Map: React.FC<MapProps> = ({ tiles, maxRow, maxColumn }) => {
    React.useEffect(() => {
        // prevent "go back/forward" on overscroll
        const prev = document.body.style.overscrollBehaviorX
        document.body.style.overscrollBehaviorX = "none"
        return () => {
            document.body.style.overscrollBehaviorX = prev
        }
    })

    const { height, width, margin } = useTiles()

    const containerRef = React.useRef<HTMLDivElement>(null)
    const [lastPosition, setLastPosition] = React.useState<Position | null>(null)
    const [offset, setOffset] = React.useState<Position>({ x: 0, y: 0 })
    const [visibility, setVisibility] = React.useState<VisibilityState>({
        minRow: 0,
        maxRow: 0,
        minColumn: 0,
        maxColumn: 0,
    })

    React.useEffect(() => {
        const container = containerRef.current
        if (!container) return

        const containerRect = container.getBoundingClientRect()

        const maxVisibleRows = Math.ceil((containerRect.height + 2 * margin.x) / height)
        const maxVisibleColumns = Math.ceil((containerRect.width + 2 * margin.y) / width)
        const skippedColumnCount = Math.ceil(-(offset.x + margin.x) / width)
        const skippedRowCount = Math.ceil(-(offset.y + margin.y) / height)
        setVisibility({
            minRow: skippedRowCount - overshot.y,
            maxRow: skippedRowCount + maxVisibleRows + overshot.y,
            minColumn: skippedColumnCount - overshot.x,
            maxColumn: skippedColumnCount + maxVisibleColumns + overshot.x,
        })
    }, [offset.x, offset.y, height, width, margin.x, margin.y])

    const mapHeight = Math.ceil(maxRow / 2) * height + (Math.floor(maxRow / 2) * height) / 2
    const mapWidth = (maxColumn + 0.5) * width

    const handlePan = (dx: number, dy: number) => {
        setOffset((prev) => {
            const container = containerRef.current
            if (!container) return prev

            const containerRect = container.getBoundingClientRect()

            const next = { x: prev.x, y: prev.y }
            if (containerRect.width < mapWidth + 2 * margin.x) {
                const maxX = mapWidth - containerRect.width + 2.5 * margin.x
                next.x = Math.max(-maxX, Math.min(0, prev.x + dx))
            } else {
                next.x = 0
            }

            if (containerRect.height < mapHeight + 2 * margin.y) {
                const maxY = mapHeight - containerRect.height + 2.75 * margin.y
                next.y = Math.max(-maxY, Math.min(0, prev.y + dy))
            } else {
                next.y = 0
            }

            return next
        })
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
                    margin: `${margin.y}px ${margin.x}px`,
                }}
            >
                {tiles.map((tile: TileProps) => (
                    <Tile
                        key={`(${tile.row},${tile.column})`}
                        {...tile}
                        visible={
                            (tile.row >= visibility.minRow && tile.row <= visibility.maxRow) ||
                            (tile.column >= visibility.minColumn &&
                                tile.column <= visibility.minColumn)
                        }
                    />
                ))}
            </div>
        </div>
    )
}

export default Map
