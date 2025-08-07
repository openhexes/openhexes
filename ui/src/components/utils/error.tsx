import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { AlertCircleIcon } from "lucide-react"
import type React from "react"

type P = React.PropsWithChildren & {
    title?: string
    error: Error
}

export const ErrorView: React.FC<P> = ({ title, error, children }) => {
    return (
        <>
            <div className="max-w-lg">
                <Alert variant="destructive" className="rounded-sm">
                    <AlertCircleIcon />
                    <AlertTitle>{title ?? error.name}</AlertTitle>
                    <AlertDescription>
                        <code className="bg-muted text-muted-foreground rounded-sm py-2 px-3 mt-2 font-mono text-sm">
                            {error.message}
                        </code>
                    </AlertDescription>
                </Alert>
            </div>
            {children}
        </>
    )
}
