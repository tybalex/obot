import {
    ClientLoaderFunctionArgs,
    redirect,
    useLoaderData,
    useNavigate,
} from "@remix-run/react";
import { useCallback } from "react";
import { $params, $path } from "remix-routes";
import { z } from "zod";

import { AgentService } from "~/lib/service/api/agentService";
import { QueryParamSchemas } from "~/lib/service/routeQueryParams";
import { noop, parseQueryParams } from "~/lib/utils";

import { Agent } from "~/components/agent";
import { Chat, ChatProvider } from "~/components/chat";
import {
    ResizableHandle,
    ResizablePanel,
    ResizablePanelGroup,
} from "~/components/ui/resizable";

export const clientLoader = async ({
    params,
    request,
}: ClientLoaderFunctionArgs) => {
    const { agent: agentId } = $params("/agents/:agent", params);
    const { threadId, from } =
        parseQueryParams(request.url, QueryParamSchemas.Agents).data || {};

    if (!agentId) {
        throw redirect("/agents");
    }

    // preload the agent
    const agent = await AgentService.getAgentById(agentId).catch(noop);

    if (!agent) {
        throw redirect("/agents");
    }
    return { agent, threadId, from };
};

export default function ChatAgent() {
    const { agent, threadId } = useLoaderData<typeof clientLoader>();
    const navigate = useNavigate();

    const updateThreadId = useCallback(
        (newThreadId?: Nullish<string>) => {
            navigate(
                $path(
                    "/agents/:agent",
                    { agent: agent.id },
                    newThreadId ? { threadId: newThreadId } : undefined
                )
            );
        },
        [agent, navigate]
    );

    return (
        <div className="h-full flex flex-col overflow-hidden relative">
            <ChatProvider
                id={agent.id}
                threadId={threadId}
                onCreateThreadId={updateThreadId}
            >
                <ResizablePanelGroup
                    direction="horizontal"
                    className="flex-auto"
                >
                    <ResizablePanel>
                        <Agent agent={agent} onRefresh={updateThreadId} />
                    </ResizablePanel>
                    <ResizableHandle withHandle />
                    <ResizablePanel>
                        <Chat />
                    </ResizablePanel>
                </ResizablePanelGroup>
            </ChatProvider>
        </div>
    );
}
