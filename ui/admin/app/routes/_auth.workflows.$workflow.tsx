import {
    ClientLoaderFunctionArgs,
    redirect,
    useLoaderData,
} from "@remix-run/react";
import { $params } from "remix-routes";

import { WorkflowService } from "~/lib/service/api/workflowService";
import { noop } from "~/lib/utils";

import { Workflow } from "~/components/workflow";

export const clientLoader = async ({ params }: ClientLoaderFunctionArgs) => {
    const { workflow: id } = $params("/workflows/:workflow", params);

    if (!id) {
        throw redirect("/threads");
    }

    const workflow = await WorkflowService.getWorkflowById(id).catch(noop);
    if (!workflow) throw redirect("/agents");

    return { workflow };
};

export default function ChatAgent() {
    const { workflow } = useLoaderData<typeof clientLoader>();

    return <Workflow workflow={workflow} />;
}
