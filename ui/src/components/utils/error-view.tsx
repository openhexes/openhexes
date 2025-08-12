import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Button } from "@/components/ui/button"
import { ConnectError } from "@connectrpc/connect"
import { AlertCircleIcon, CopyIcon } from "lucide-react"
import type React from "react"

type P = React.PropsWithChildren & {
    title?: string
    error: Error
}

function isConnectError(err: Error): err is ConnectError {
    const e = err as unknown as ConnectError
    if (e.code !== undefined) {
        return true
    }
    return false
}

export const ErrorView: React.FC<P> = ({ title, error, children }) => {
    title = title ?? error.name ?? "Unknown error"
    if (isConnectError(error)) {
        title = `Error [code=${error.code}]`
    }

    const copyError = async () => {
        const errorText = `${title}: ${error.message}`
        await navigator.clipboard.writeText(errorText)
    }

    return (
        <div className="h-screen w-screen flex items-center justify-center">
            <div className="w-lg">
                <Alert className="rounded-sm p-2">
                    <AlertTitle className="flex items-center gap-2 text-destructive mb-2">
                        <AlertCircleIcon size={16} />
                        {title}
                    </AlertTitle>
                    <AlertDescription className="flex flex-col gap-2 truncate">
                        <code className="border-1 bg-muted text-muted-foreground rounded-sm py-2 px-3 font-mono text-sm w-full whitespace-pre overflow-scroll max-h-lg">
                            {error.message}
                        </code>
                        <div className="flex gap-2 items-center">
                            <Button size="sm" variant="outline" onClick={void copyError}>
                                <CopyIcon />
                                Copy
                            </Button>
                            {children}
                        </div>
                    </AlertDescription>
                </Alert>
            </div>
        </div>
    )
}
