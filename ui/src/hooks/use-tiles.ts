import React from "react"

export const useTiles = () => {
    const [height, setHeight] = React.useState(64)
    const width = height * Math.cos(Math.PI / 6)
    const margin = {
        x: 2 * width,
        y: 2 * height * 0.75,
    }
    return { height, width, margin, setHeight }
}
