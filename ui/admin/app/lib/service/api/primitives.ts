// TODO: Add default configurations with auth tokens, etc. When ready
import axios, { AxiosRequestConfig, AxiosResponse } from "axios";

export const ResponseHeaders = {
    RunId: "x-otto-run-id",
    ThreadId: "x-otto-thread-id",
} as const;

const internalFetch = axios.request;

interface ExtendedAxiosRequestConfig<D = unknown>
    extends AxiosRequestConfig<D> {
    errorMessage?: string;
}

export async function request<T, R = AxiosResponse<T>, D = unknown>({
    errorMessage = "Request failed",
    ...config
}: ExtendedAxiosRequestConfig<D>): Promise<R> {
    console.error(errorMessage);
    return await internalFetch<T, R, D>({
        adapter: "fetch",
        ...config,
    });
}
