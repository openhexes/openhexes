import { createContext, useContext } from "react"

export type Theme = "dark" | "light" | "system"

type S = {
    theme: Theme
    setTheme: (theme: Theme) => void
}

export const ThemeProviderContext = createContext<S>({
    theme: "dark",
    setTheme: () => null,
})

export const useTheme = () => {
    const context = useContext(ThemeProviderContext)

    if (context === undefined) throw new Error("useTheme must be used within a ThemeProvider")

    return context
}
