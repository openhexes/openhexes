import { queryClient } from "@/hooks/client"
import { useEnv } from "@/hooks/use-env"
import { GoogleOAuthProvider } from "@react-oauth/google"
import { QueryClientProvider } from "@tanstack/react-query"
import { ReactQueryDevtools } from "@tanstack/react-query-devtools"
import { NuqsAdapter } from "nuqs/adapters/react-router/v7"
import React from "react"
import { BrowserRouter, Routes } from "react-router"

import { AuthWall } from "./auth-wall"

export const Layout: React.FC<React.PropsWithChildren> = ({ children }) => {
    const googleClientID = useEnv("GOOGLE_CLIENT_ID")

    return (
        <GoogleOAuthProvider clientId={googleClientID}>
            <QueryClientProvider client={queryClient}>
                <NuqsAdapter>
                    <AuthWall>
                        <BrowserRouter>
                            <Routes>{children}</Routes>
                        </BrowserRouter>
                    </AuthWall>
                    <ReactQueryDevtools initialIsOpen={false} />
                </NuqsAdapter>
            </QueryClientProvider>
        </GoogleOAuthProvider>
    )
}

export default Layout
