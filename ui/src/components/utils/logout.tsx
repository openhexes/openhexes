import { Button } from "@/components/ui/button"
import { cookieName } from "@/lib/const"
import { googleLogout } from "@react-oauth/google"
import Cookies from "js-cookie"
import { LogOut } from "lucide-react"
import React from "react"

const logOut = (reload = true) => {
    Cookies.remove(cookieName, { path: "/", sameSite: "strict" })
    googleLogout()

    if (reload) {
        document.location.reload()
    }
}

export const LogOutButton: React.FC = () => {
    return (
        <Button variant="outline" onClick={() => logOut()}>
            <LogOut />
            Log out
        </Button>
    )
}
