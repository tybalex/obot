// TODO: Add default configurations with auth tokens, etc. When ready
import axios, { AxiosRequestConfig, AxiosResponse } from "axios";
import { toast } from "sonner";

export const ResponseHeaders = {
    RunId: "x-otto-run-id",
    ThreadId: "x-otto-thread-id",
} as const;

const internalFetch = axios.request;

interface ExtendedAxiosRequestConfig<D = unknown> extends AxiosRequestConfig<D> {
    throwErrors?: boolean;
    errorMessage?: string;
}

export async function request<T, R = AxiosResponse<T>, D = unknown>({
    throwErrors = true,
    errorMessage = "Request failed",
    ...config
}: ExtendedAxiosRequestConfig<D>): Promise<R> {
    try {
        return await internalFetch<T, R, D>({
            adapter: "fetch",
            ...config,
        });
    } catch (error) {
        handleRequestError(error, errorMessage);
        
        if (throwErrors) {
            throw error;
        }
        return error as R;
    }
}

function handleRequestError(error: unknown, errorMessage: string): void {
    if (axios.isAxiosError(error) && error.response) {
        const { status, config } = error.response;
        const method = config.method?.toUpperCase() || 'UNKNOWN';
        toast.error(`${status} ${method}`, {
            description: errorMessage,
        });
    } else {
        toast.error("Request Error", {
            description: errorMessage,
        });
    }
}