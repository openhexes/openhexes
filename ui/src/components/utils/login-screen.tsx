import { useTitle } from "@/hooks/use-title"
import { cookieName } from "@/lib/const"
import { GoogleLogin, googleLogout } from "@react-oauth/google"
import Cookies from "js-cookie"
import React from "react"
import { toast } from "sonner"

import { LogOutButton } from "./logout"

const logOut = (reload = true) => {
    Cookies.remove(cookieName, { path: "/", sameSite: "strict" })
    googleLogout()

    if (reload) {
        document.location.reload()
    }
}

// todo: default button

export const LoginScreen: React.FC = () => {
    useTitle("Login")

    return (
        <div className="flex items-center justify-center h-screen flex-col gap-4">
            <GoogleLogin
                type="icon"
                shape="circle"
                onSuccess={(credentialResponse) => {
                    const credential = credentialResponse.credential ?? ""
                    Cookies.set(cookieName, credential, { path: "/", sameSite: "strict" })
                }}
                onError={() => {
                    toast.error("login failed")
                    logOut(false)
                }}
            />
            <LogOutButton />
        </div>
    )
}
