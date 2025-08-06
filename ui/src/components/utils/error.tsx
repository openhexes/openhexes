import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Card, CardContent } from "@/components/ui/card"
import { AlertCircleIcon } from "lucide-react"
import type React from "react"

type P = React.PropsWithChildren & {
    title?: string
    error: Error
}

export const ErrorView: React.FC<P> = ({ title, error, children }) => {
    return (
        <Card className="min-w-[340px] w-full max-w-md">
            <CardContent>
                <div className="grid gap-6">
                    <div className="flex flex-col gap-4">
                        <Alert variant="destructive">
                            <AlertCircleIcon />
                            <AlertTitle>{title ?? error.name}</AlertTitle>
                            <AlertDescription>
                                <p>Please try again later.</p>
                                <code className="bg-muted relative rounded px-[0.3rem] py-[0.2rem] font-mono text-sm font-semibold">
                                    {error.message}
                                </code>
                            </AlertDescription>
                        </Alert>
                        {children}
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}
