import { useEffect } from "react"

export const useTitle = (subtitle: string, base: string = "Hexes | ") => {
    useEffect(() => {
        document.title = `${base}${subtitle}`
        console.warn(`document title -> "${document.title}"`)
    }, [base, subtitle])
}
