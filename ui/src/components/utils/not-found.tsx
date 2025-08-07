import { Button } from "@/components/ui/button"
import { useTitle } from "@/hooks/use-title"
import { Home } from "lucide-react"
import React from "react"
import { Link, useLocation } from "react-router"

import { LogOutButton } from "./logout"

export const NotFound: React.FC = () => {
    useTitle("Page not found")

    const location = useLocation()

    return (
        <>
            <div className="p-4 flex flex-col gap-4">
                <h2>Requested page doesn't exist:</h2>
                <code className="bg-muted p-4 rounded-md">{location.pathname}</code>
                <div className="flex items-center gap-2">
                    <Link to="/">
                        <Button variant="outline">
                            <Home />
                            Go to home page
                        </Button>
                    </Link>
                    <LogOutButton />
                </div>
            </div>
        </>
    )
}

export default NotFound
