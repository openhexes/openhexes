import { useCurrentAccount } from "@/hooks/fetch"
import { cookieName } from "@/lib/const"
import Cookies from "js-cookie"
import { Loader2 } from "lucide-react"
import React from "react"

import { ErrorView } from "./error-view"
import { LoginScreen } from "./login-screen"
import { LogOutButton } from "./logout"

export const AuthWall: React.FC<React.PropsWithChildren> = ({ children }) => {
    const account = useCurrentAccount()

    if (account.isSuccess) {
        return children
    }
    if (account.isLoading) {
        return (
            <div className="flex items-center justify-center h-screen">
                <Loader2 className="animate-spin h-8 w-8" />
            </div>
        )
    }

    if (!Cookies.get(cookieName)) {
        return <LoginScreen />
    }

    return (
        <ErrorView error={account.error ?? new Error("unknown error")}>
            <LogOutButton size="sm" />
        </ErrorView>
    )
}
