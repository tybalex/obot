import { PersonIcon, ReaderIcon } from "@radix-ui/react-icons";
import {
    ClientLoaderFunctionArgs,
    Link,
    useLoaderData,
    useNavigate,
} from "@remix-run/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { PuzzleIcon, Trash } from "lucide-react";
import { useMemo } from "react";
import { $path } from "remix-routes";
import useSWR, { preload } from "swr";
import { z } from "zod";

import { Agent } from "~/lib/model/agents";
import { Thread } from "~/lib/model/threads";
import { Workflow } from "~/lib/model/workflows";
import { AgentService } from "~/lib/service/api/agentService";
import { ThreadsService } from "~/lib/service/api/threadsService";
import { WorkflowService } from "~/lib/service/api/workflowService";
import { RouteService } from "~/lib/service/routeQueryParams";
import { timeSince } from "~/lib/utils";

import { TypographyH2, TypographyP } from "~/components/Typography";
import { DataTable } from "~/components/composed/DataTable";
import { Button } from "~/components/ui/button";
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "~/components/ui/tooltip";
import { useAsync } from "~/hooks/useAsync";

export type SearchParams = z.infer<(typeof RouteService.schemas)["/threads"]>;

export async function clientLoader({ request }: ClientLoaderFunctionArgs) {
    const search = new URL(request.url).search;

    await Promise.all([
        preload(AgentService.getAgents.key(), AgentService.getAgents),
        preload(
            WorkflowService.getWorkflows.key(),
            WorkflowService.getWorkflows
        ),
        preload(ThreadsService.getThreads.key(), ThreadsService.getThreads),
    ]);

    return RouteService.getQueryParams("/threads", search) ?? {};
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

    const getWorkflows = useSWR(
        WorkflowService.getWorkflows.key(),
        WorkflowService.getWorkflows
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

    const workflowMap = useMemo(() => {
        if (!getWorkflows.data || !Array.isArray(getWorkflows.data)) return {};
        return getWorkflows.data.reduce(
            (acc, workflow) => {
                acc[workflow.id] = workflow;
                return acc;
            },
            {} as Record<string, Workflow>
        );
    }, [getWorkflows.data]);

    const threads = useMemo(() => {
        console.log(agentId);
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
                    data={threads.filter(
                        (thread) => thread.agentID || thread.workflowID
                    )}
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
                    else if (thread.workflowID)
                        return (
                            workflowMap[thread.workflowID]?.name ??
                            thread.workflowID
                        );
                    return "Unnamed";
                },
                { header: "Name" }
            ),
            columnHelper.display({
                id: "type",
                header: "Type",
                cell: ({ row }) => {
                    return (
                        <TypographyP className="flex items-center gap-2">
                            {row.original.agentID ? (
                                <PersonIcon className="w-4 h-4" />
                            ) : (
                                <PuzzleIcon className="w-4 h-4" />
                            )}
                            {row.original.agentID ? "Agent" : "Workflow"}
                        </TypographyP>
                    );
                },
            }),
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
                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button variant="ghost" size="icon" asChild>
                                    <Link
                                        to={$path("/thread/:id", {
                                            id: row.original.id,
                                        })}
                                    >
                                        <ReaderIcon width={21} height={21} />
                                    </Link>
                                </Button>
                            </TooltipTrigger>

                            <TooltipContent>
                                <p>Inspect Thread</p>
                            </TooltipContent>
                        </Tooltip>

                        <Tooltip>
                            <TooltipTrigger asChild>
                                <Button
                                    variant="ghost"
                                    size="icon"
                                    onClick={() =>
                                        deleteThread.execute(row.original.id)
                                    }
                                >
                                    <Trash />
                                </Button>
                            </TooltipTrigger>

                            <TooltipContent>
                                <p>Delete Thread</p>
                            </TooltipContent>
                        </Tooltip>
                    </div>
                ),
            }),
        ];
    }
}

const columnHelper = createColumnHelper<Thread>();
