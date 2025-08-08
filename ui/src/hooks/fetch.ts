import { cookieName } from "@/lib/const"
import { create, toJson } from "@bufbuild/protobuf"
import { createClient } from "@connectrpc/connect"
import { createGrpcWebTransport } from "@connectrpc/connect-web"
import { useQuery } from "@tanstack/react-query"
import Cookies from "js-cookie"
import { GameService } from "proto/ts/game/v1/game_pb"
import {
    type Account,
    IAMService,
    type ListAccountsRequest,
    ListAccountsRequestSchema,
} from "proto/ts/iam/v1/iam_pb"
import { toast } from "sonner"

const noCookieErrorMessage = "auth cookie not set"
const invalidArgumentMessage = "[invalid_argument]"

const transport = createGrpcWebTransport({
    baseUrl: (import.meta.env.VITE_API_ADDRESS as string) || "http://localhost:8080",
    useBinaryFormat: true, // switch to false to use JSON, bodies will be readable in devtools
    defaultTimeoutMs: 30000,
    fetch: (input, init) => fetch(input, { ...init, credentials: "include" }),
})

export const IAMClient = createClient(IAMService, transport)
export const GameClient = createClient(GameService, transport)

const handleError =
    (op: string, maxAttempts = 3) =>
    (failureCount: number, error: Error): boolean => {
        toast.error(op, { description: error.message, closeButton: true })
        if (
            error.message.includes(noCookieErrorMessage) ||
            error.message.includes("[unauthenticated]") ||
            error.message.includes("[permission_denied]") ||
            error.message.includes(invalidArgumentMessage) ||
            error.message.includes("[not_found]")
        ) {
            return false
        }
        return failureCount < maxAttempts
    }

export const useCurrentAccount = () => {
    return useQuery({
        queryKey: ["accounts", "self"],
        queryFn: async ({ signal }) => {
            if (!Cookies.get(cookieName)) {
                throw new Error(noCookieErrorMessage)
            }
            return (await IAMClient.resolveAccount({}, { signal })).account
        },
        retry: handleError("Authentication failed"),
    })
}

export const useAccounts = (request: ListAccountsRequest) => {
    request = create(ListAccountsRequestSchema, request)

    return useQuery({
        queryKey: ["accounts", toJson(ListAccountsRequestSchema, request)],
        queryFn: async ({ signal }) => {
            const accounts: Account[] = []
            for await (const chunk of IAMClient.listAccounts(request, { signal })) {
                accounts.push(...chunk.accounts)
            }
            return accounts
        },
        retry: handleError("Failed to fetch accounts"),
    })
}
