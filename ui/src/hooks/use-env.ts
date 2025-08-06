export const useEnv = (key: string) => {
    const value = import.meta.env[key] as string | undefined
    return value || ""
}
