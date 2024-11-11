import { List, PuzzleIcon, WrenchIcon } from "lucide-react";
import { useCallback, useState } from "react";

import { Workflow as WorkflowType } from "~/lib/model/workflows";
import { cn } from "~/lib/utils";

import { TypographyH4, TypographyP } from "~/components/Typography";
import { AgentForm } from "~/components/agent";
import { Card, CardDescription } from "~/components/ui/card";
import { ScrollArea } from "~/components/ui/scroll-area";
import { BasicToolForm } from "~/components/workflow/BasicToolForm";
import { ParamsForm } from "~/components/workflow/ParamsForm";
import {
    WorkflowProvider,
    useWorkflow,
} from "~/components/workflow/WorkflowContext";
import { StepsForm } from "~/components/workflow/steps/StepsForm";
import { useDebounce } from "~/hooks/useDebounce";

type WorkflowProps = {
    workflow: WorkflowType;
    className?: string;
};

export function Workflow(props: WorkflowProps) {
    return (
        <WorkflowProvider workflow={props.workflow}>
            <WorkflowContent {...props} />
        </WorkflowProvider>
    );
}

function WorkflowContent({ className }: WorkflowProps) {
    const { workflow, updateWorkflow, isUpdating, lastUpdated } = useWorkflow();

    const [workflowUpdates, setWorkflowUpdates] = useState(workflow);

    const partialSetWorkflow = useCallback(
        (changes: Partial<typeof workflow>) => {
            console.log("called");

            const updatedWorkflow = {
                ...workflow,
                ...workflowUpdates,
                ...changes,
            };

            updateWorkflow(updatedWorkflow);

            setWorkflowUpdates(updatedWorkflow);
        },
        [updateWorkflow, workflow, workflowUpdates]
    );

    const debouncedSetWorkflowInfo = useDebounce(partialSetWorkflow, 1000);

    return (
        <div className="h-full flex flex-col">
            <ScrollArea className={cn("h-full", className)}>
                <Card className="p-4 m-4">
                    <AgentForm
                        agent={workflowUpdates}
                        onChange={debouncedSetWorkflowInfo}
                    />
                </Card>

                <Card className="p-4 m-4 flex flex-col gap-4">
                    <TypographyH4 className="flex items-center gap-2">
                        <WrenchIcon className="w-5 h-5" />
                        Tools
                    </TypographyH4>

                    <CardDescription>
                        Add tools the allow the agent to perform useful actions
                        such as searching the web, reading files, or interacting
                        with other systems.
                    </CardDescription>

                    <BasicToolForm
                        defaultValues={workflow}
                        onChange={debouncedSetWorkflowInfo}
                    />
                </Card>

                <Card className="p-4 m-4 flex flex-col gap-4">
                    <TypographyH4 className="flex items-center gap-2">
                        <List className="w-4 h-4" />
                        Parameters
                    </TypographyH4>

                    <ParamsForm
                        workflow={workflow}
                        onChange={(values) =>
                            debouncedSetWorkflowInfo({
                                params: values.params,
                            })
                        }
                    />
                </Card>

                <Card className="p-4 m-4 flex flex-col gap-4">
                    <TypographyH4 className="flex items-center gap-2">
                        <PuzzleIcon className="w-4 h-4" />
                        Steps
                    </TypographyH4>

                    <StepsForm
                        workflow={workflow}
                        onChange={(values) =>
                            debouncedSetWorkflowInfo({ steps: values.steps })
                        }
                    />
                </Card>
            </ScrollArea>

            <footer className="flex justify-between items-center p-4 gap-4 text-muted-foreground">
                {isUpdating ? (
                    <TypographyP>Saving...</TypographyP>
                ) : lastUpdated ? (
                    <TypographyP>Saved</TypographyP>
                ) : (
                    <div />
                )}
            </footer>
        </div>
    );
}
