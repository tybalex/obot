import { ReaderIcon } from "@radix-ui/react-icons";
import {
    ClientLoaderFunctionArgs,
    Link,
    useLoaderData,
    useNavigate,
} from "@remix-run/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Trash } from "lucide-react";
import { useMemo } from "react";
import { $path } from "remix-routes";
import useSWR from "swr";
import { z } from "zod";

import { Agent } from "~/lib/model/agents";
import { Thread } from "~/lib/model/threads";
import { AgentService } from "~/lib/service/api/agentService";
import { ThreadsService } from "~/lib/service/api/threadsService";
import { QueryParamSchemas } from "~/lib/service/routeQueryParams";
import { parseQueryParams, timeSince } from "~/lib/utils";

import { TypographyH2, TypographyP } from "~/components/Typography";
import { DataTable } from "~/components/composed/DataTable";
import { Button } from "~/components/ui/button";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "~/components/ui/tooltip";
import { useAsync } from "~/hooks/useAsync";

const paramSchema = QueryParamSchemas.Threads;

export type SearchParams = z.infer<typeof paramSchema>;

export function clientLoader({ request }: ClientLoaderFunctionArgs) {
    return parseQueryParams(request.url, paramSchema).data || {};
}

export default function Threads() {
    const navigate = useNavigate();
    const { agentId, workflowId } = useLoaderData<typeof clientLoader>();

    const getThreads = useSWR(
        ThreadsService.getThreads.key(),
        ThreadsService.getThreads
    );

    const getAgents = useSWR(
        AgentService.getAgents.key(),
        AgentService.getAgents
    );

    const agentMap = useMemo(() => {
        // note(tylerslaton): the or condition here is because the getAgents.data can
        // be an object containing a url only when switching to the agent page from the
        // threads page.
        if (!getAgents.data || !Array.isArray(getAgents.data)) return {};
        return getAgents.data.reduce(
            (acc, agent) => {
                acc[agent.id] = agent;
                return acc;
            },
            {} as Record<string, Agent>
        );
    }, [getAgents.data]);

    const threads = useMemo(() => {
        if (!getThreads.data) return [];

        if (!agentId && !workflowId) return getThreads.data;

        switch (true) {
            case !!agentId:
                return getThreads.data.filter(
                    (thread) => thread.agentID === agentId
                );
            case !!workflowId:
                return getThreads.data.filter(
                    (thread) => thread.workflowID === workflowId
                );
            default:
                return getThreads.data;
        }
    }, [getThreads.data, agentId, workflowId]);

    const deleteThread = useAsync(ThreadsService.deleteThread, {
        onSuccess: ThreadsService.revalidateThreads,
    });

    return (
        <div className="h-full flex flex-col">
            <div className="flex-auto overflow-hidden p-8">
                <TypographyH2 className="mb-8">Threads</TypographyH2>
                <DataTable
                    columns={getColumns()}
                    data={threads.filter((thread) => thread.agentID)}
                    sort={[{ id: "created", desc: true }]}
                    classNames={{
                        row: "!max-h-[200px] grow-0 height-[200px]",
                        cell: "!max-h-[200px] grow-0 height-[200px]",
                    }}
                    disableClickPropagation={(cell) =>
                        cell.id.includes("actions")
                    }
                    onRowClick={(row) => {
                        navigate($path("/thread/:id", { id: row.id }));
                    }}
                />
            </div>
        </div>
    );

    function getColumns(): ColumnDef<Thread, string>[] {
        return [
            columnHelper.accessor(
                (thread) => {
                    if (thread.agentID)
                        return agentMap[thread.agentID]?.name ?? thread.agentID;
                    return "Unnamed";
                },
                { header: "Agent" }
            ),
            columnHelper.accessor("created", {
                id: "created",
                header: "Created",
                cell: (info) => (
                    <TypographyP>
                        {timeSince(new Date(info.row.original.created))} ago
                    </TypographyP>
                ),
                sortingFn: "datetime",
            }),
            columnHelper.display({
                id: "actions",
                cell: ({ row }) => (
                    <div className="flex gap-2 justify-end">
                        <TooltipProvider>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button variant="ghost" size="icon" asChild>
                                        <Link
                                            to={$path("/thread/:id", {
                                                id: row.original.id,
                                            })}
                                        >
                                            <ReaderIcon
                                                width={21}
                                                height={21}
                                            />
                                        </Link>
                                    </Button>
                                </TooltipTrigger>

                                <TooltipContent>
                                    <p>Inspect Thread</p>
                                </TooltipContent>
                            </Tooltip>
                        </TooltipProvider>

                        <TooltipProvider>
                            <Tooltip>
                                <TooltipTrigger asChild>
                                    <Button
                                        variant="ghost"
                                        size="icon"
                                        onClick={() =>
                                            deleteThread.execute(
                                                row.original.id
                                            )
                                        }
                                    >
                                        <Trash />
                                    </Button>
                                </TooltipTrigger>

                                <TooltipContent>
                                    <p>Delete Thread</p>
                                </TooltipContent>
                            </Tooltip>
                        </TooltipProvider>
                    </div>
                ),
            }),
        ];
    }
}

const columnHelper = createColumnHelper<Thread>();
