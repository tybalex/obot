import queryString from "query-string";
import { $params, $path, Routes, RoutesWithParams } from "remix-routes";
import { ZodSchema, z } from "zod";

const QueryParamSchemaMap = {
    "": z.undefined(),
    "/": z.undefined(),
    "/agents": z.undefined(),
    "/agents/:agent": z.object({
        threadId: z.string().optional(),
        from: z.string().optional(),
    }),
    "/debug": z.undefined(),
    "/home": z.undefined(),
    "/oauth-apps": z.undefined(),
    "/thread/:id": z.undefined(),
    "/threads": z.object({
        agentId: z.string().optional(),
        workflowId: z.string().optional(),
    }),
    "/workflows": z.undefined(),
    "/workflows/:workflow": z.undefined(),
    "/tools": z.undefined(),
    "/users": z.undefined(),
} satisfies Record<keyof Routes, ZodSchema | null>;

function parseSearchParams<T extends keyof Routes>(route: T, search: string) {
    if (!QueryParamSchemaMap[route])
        throw new Error(`No schema found for route: ${route}`);

    const obj = queryString.parse(search);
    const { data, success } = QueryParamSchemaMap[route].safeParse(obj);

    if (!success) {
        console.error("Failed to parse query params", route, search);
        return undefined;
    }

    return data as z.infer<(typeof QueryParamSchemaMap)[T]>;
}

type QueryParamInfo<T extends keyof Routes> = {
    path: T;
    query?: z.infer<(typeof QueryParamSchemaMap)[T]>;
};

function getUnknownQueryParams(pathname: string, search: string) {
    if (new RegExp($path("/agents/:agent", { agent: "(.*)" })).test(pathname)) {
        return {
            path: "/agents/:agent",
            query: parseSearchParams("/agents/:agent", search),
        } satisfies QueryParamInfo<"/agents/:agent">;
    }

    if (new RegExp($path("/threads")).test(pathname)) {
        return {
            path: "/threads",
            query: parseSearchParams("/threads", search),
        } satisfies QueryParamInfo<"/threads">;
    }

    if (
        new RegExp($path("/workflows/:workflow", { workflow: "(.*)" })).test(
            pathname
        )
    ) {
        return {
            path: "/workflows/:workflow",
            query: parseSearchParams("/workflows/:workflow", search),
        } satisfies QueryParamInfo<"/workflows/:workflow">;
    }

    return {};
}

type PathParamInfo<T extends keyof RoutesWithParams> = {
    path: T;
    pathParams: ReturnType<typeof $params<T, Routes[T]["params"]>>;
};

function getUnknownPathParams(
    pathname: string,
    params: Record<string, string | undefined>
) {
    if (new RegExp($path("/agents/:agent", { agent: "(.*)" })).test(pathname)) {
        return {
            path: "/agents/:agent",
            pathParams: $params("/agents/:agent", params),
        } satisfies PathParamInfo<"/agents/:agent">;
    }

    if (new RegExp($path("/thread/:id", { id: "(.*)" })).test(pathname)) {
        return {
            path: "/thread/:id",
            pathParams: $params("/thread/:id", params),
        } satisfies PathParamInfo<"/thread/:id">;
    }

    return {};
}

export const RouteService = {
    getPathParams: $params,
    getUnknownPathParams,
    getUnknownQueryParams,
    getQueryParams: parseSearchParams,
    schemas: QueryParamSchemaMap,
};
