import React from "react"

// prevents "go back/forward" and bounce on overscroll
export const useNoOverscroll = () => {
    return React.useEffect(() => {
        const prevX = document.body.style.overscrollBehaviorX
        const prevY = document.body.style.overscrollBehaviorY

        // Prevent both horizontal and vertical overscroll
        document.body.style.overscrollBehaviorX = "none"
        document.body.style.overscrollBehaviorY = "none"

        // Additional wheel event prevention for aggressive overscroll blocking
        const preventWheel = (e: WheelEvent) => {
            // Only prevent default if we're at the edge of scrollable content
            const target = e.target as Element
            const mapContainer = target.closest('[data-testid="map-container"]')
            if (mapContainer) {
                e.preventDefault()
            }
        }

        document.addEventListener("wheel", preventWheel, { passive: false })

        return () => {
            document.body.style.overscrollBehaviorX = prevX
            document.body.style.overscrollBehaviorY = prevY
            document.removeEventListener("wheel", preventWheel)
        }
    }, [])
}
