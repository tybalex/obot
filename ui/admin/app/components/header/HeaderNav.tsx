import { Link, Params, useLocation, useParams } from "@remix-run/react";
import { $path } from "remix-routes";
import useSWR from "swr";

import { AgentService } from "~/lib/service/api/agentService";
import { ThreadsService } from "~/lib/service/api/threadsService";
import { cn } from "~/lib/utils";

import { DarkModeToggle } from "~/components/DarkModeToggle";
import {
    Breadcrumb,
    BreadcrumbItem,
    BreadcrumbLink,
    BreadcrumbList,
    BreadcrumbPage,
    BreadcrumbSeparator,
} from "~/components/ui/breadcrumb";
import { SidebarTrigger } from "~/components/ui/sidebar";
import { UserMenu } from "~/components/user/UserMenu";

export function HeaderNav() {
    const { pathname } = useLocation();
    const params = useParams();
    const headerHeight = "h-[60px]";

    return (
        <header
            className={cn(
                "flex transition-all duration-300 ease-in-out",
                headerHeight
            )}
        >
            <div className="h-full flex-auto flex">
                <div className="flex flex-grow border-b">
                    <div className="flex-grow flex justify-start items-center p-4">
                        <SidebarTrigger className="h-4 w-4" />
                        <div className="border-r h-4 mx-4" />
                        {getBreadcrumbs(pathname, params)}
                    </div>

                    <div className="flex items-center justify-center p-4 mr-4">
                        <UserMenu className="pr-4 border-r mr-4" />
                        <DarkModeToggle />
                    </div>
                </div>
            </div>
        </header>
    );
}

function getBreadcrumbs(route: string, params: Readonly<Params<string>>) {
    return (
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <BreadcrumbLink asChild>
                        <Link to={$path("/")}>Home</Link>
                    </BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                {new RegExp($path("/agents/:agent", { agent: "(.*)" })).test(
                    route
                ) ? (
                    <>
                        <BreadcrumbItem>
                            <BreadcrumbLink asChild>
                                <Link to={$path("/agents")}>Agents</Link>
                            </BreadcrumbLink>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                            <BreadcrumbPage>
                                <AgentName agentId={params.agent || ""} />
                            </BreadcrumbPage>
                        </BreadcrumbItem>
                    </>
                ) : (
                    new RegExp($path("/agents")).test(route) && (
                        <BreadcrumbItem>
                            <BreadcrumbPage>Agents</BreadcrumbPage>
                        </BreadcrumbItem>
                    )
                )}
                {new RegExp($path("/threads")).test(route) && (
                    <BreadcrumbItem>
                        <BreadcrumbPage>Threads</BreadcrumbPage>
                    </BreadcrumbItem>
                )}
                {new RegExp($path("/thread/:id", { id: "(.*)" })).test(
                    route
                ) && (
                    <>
                        <BreadcrumbItem>
                            <BreadcrumbLink asChild>
                                <Link to={$path("/threads")}>Threads</Link>
                            </BreadcrumbLink>
                        </BreadcrumbItem>
                        <BreadcrumbSeparator />
                        <BreadcrumbItem>
                            <BreadcrumbPage>
                                <ThreadName threadId={params.id || ""} />
                            </BreadcrumbPage>
                        </BreadcrumbItem>
                    </>
                )}
                {new RegExp($path("/tools")).test(route) && (
                    <BreadcrumbItem>
                        <BreadcrumbPage>Tools</BreadcrumbPage>
                    </BreadcrumbItem>
                )}
                {new RegExp($path("/users")).test(route) && (
                    <BreadcrumbItem>
                        <BreadcrumbPage>Users</BreadcrumbPage>
                    </BreadcrumbItem>
                )}
                {new RegExp($path("/oauth-apps")).test(route) && (
                    <BreadcrumbItem>
                        <BreadcrumbPage>OAuth Apps</BreadcrumbPage>
                    </BreadcrumbItem>
                )}
            </BreadcrumbList>
        </Breadcrumb>
    );
}

const AgentName = ({ agentId }: { agentId: string }) => {
    const { data: agent } = useSWR(
        AgentService.getAgentById.key(agentId),
        ({ agentId }) => AgentService.getAgentById(agentId)
    );

    return <>{agent?.name || "New Agent"}</>;
};

const ThreadName = ({ threadId }: { threadId: string }) => {
    const { data: thread } = useSWR(
        ThreadsService.getThreadById.key(threadId),
        ({ threadId }) => ThreadsService.getThreadById(threadId)
    );

    return <>{thread?.description || threadId}</>;
};
