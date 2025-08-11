import React from "react"

// prevents "go back/forward" on overscroll
export const useNoOverscroll = () => {
    return React.useEffect(() => {
        const prev = document.body.style.overscrollBehaviorX
        document.body.style.overscrollBehaviorX = "none"
        return () => {
            document.body.style.overscrollBehaviorX = prev
        }
    }, [])
}
