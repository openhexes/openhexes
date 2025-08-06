import { useCurrentAccount } from "@/hooks/fetch"
import { cookieName } from "@/lib/const"
import { GoogleLogin, googleLogout } from "@react-oauth/google"
import Cookies from "js-cookie"
import { Loader2, LogOut } from "lucide-react"
import React from "react"
import { toast } from "sonner"

import { Button } from "../ui/button"
import { ErrorView } from "./error"

const logOut = (reload = true) => {
    Cookies.remove(cookieName, { path: "/", sameSite: "strict" })
    googleLogout()

    if (reload) {
        document.location.reload()
    }
}

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
        return (
            <div className="flex items-center justify-center h-screen flex-col gap-4">
                <GoogleLogin
                    onSuccess={(credentialResponse) => {
                        const credential = credentialResponse.credential ?? ""
                        Cookies.set(cookieName, credential, { path: "/", sameSite: "strict" })
                    }}
                    onError={() => {
                        toast.error("login failed")
                        logOut(false)
                    }}
                />
                <Button variant="outline" onClick={() => logOut()}>
                    <LogOut />
                    Log out
                </Button>
            </div>
        )
    }

    return (
        <div className="flex items-center justify-center h-screen flex-col gap-4">
            <ErrorView error={account.error ?? new Error("unknown error")} />
            <Button variant="outline" onClick={() => logOut()}>
                <LogOut />
                Log out
            </Button>
        </div>
    )
}
