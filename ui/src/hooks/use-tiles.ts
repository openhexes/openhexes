import { useWindowSize } from "@uidotdev/usehooks"

export interface TileDimensions {
    height: number
    width: number
    margin: {
        x: number
        y: number
    }
}

export const useTiles = (minVisibleRows = 20): TileDimensions => {
    const screen = useWindowSize()
    const height = Math.ceil((screen.height || 600) / minVisibleRows)

    const width = height * Math.cos(Math.PI / 6)
    const margin = {
        x: 2 * width,
        y: 2 * height * 0.75,
    }

    return { height, width, margin }
}
