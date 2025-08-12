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

type P = Partial<React.ComponentProps<typeof Button>>

export const LogOutButton: React.FC<P> = (props) => {
    if (Cookies.get(cookieName) === undefined) {
        return null
    }

    return (
        <Button variant="outline" onClick={() => logOut()} {...props}>
            <LogOut />
            Log out
        </Button>
    )
}
