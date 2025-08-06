import { useEffect } from "react"

export const useTitle = (subtitle: string, base: string = "Hexes | ") => {
    useEffect(() => {
        document.title = `${base}${subtitle}`
    }, [base, subtitle])
}
