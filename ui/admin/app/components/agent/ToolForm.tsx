import { zodResolver } from "@hookform/resolvers/zod";
import { ArrowDownIcon, ArrowUpIcon, PlusIcon } from "lucide-react";
import { useEffect, useMemo } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { z } from "zod";

import { Agent } from "~/lib/model/agents";
import { noop } from "~/lib/utils";

import { ToolEntry } from "~/components/agent/ToolEntry";
import { ToolCatalog } from "~/components/tools/ToolCatalog";
import { Button } from "~/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
} from "~/components/ui/dialog";
import { Form, FormLabel } from "~/components/ui/form";

import { TypographyH4, TypographySmall } from "../Typography";
import { Switch } from "../ui/switch";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "../ui/tooltip";

const ToolVariant = {
    FIXED: "fixed",
    DEFAULT: "default",
    AVAILABLE: "available",
} as const;
type ToolVariant = (typeof ToolVariant)[keyof typeof ToolVariant];
const formSchema = z.object({
    tools: z.array(
        z.object({
            tool: z.string(),
            variant: z.enum([
                ToolVariant.FIXED,
                ToolVariant.DEFAULT,
                ToolVariant.AVAILABLE,
            ] as const),
        })
    ),
});

export type ToolFormValues = z.infer<typeof formSchema>;

export function ToolForm({
    agent,
    onSubmit,
    onChange,
}: {
    agent: Agent;
    onSubmit?: (values: ToolFormValues) => void;
    onChange?: (values: ToolFormValues) => void;
}) {
    const defaultValues = useMemo(() => {
        return {
            tools: [
                ...(agent.tools ?? []).map((tool) => ({
                    tool,
                    variant: ToolVariant.FIXED,
                })),
                ...(agent.defaultThreadTools ?? []).map((tool) => ({
                    tool,
                    variant: ToolVariant.DEFAULT,
                })),
                ...(agent.availableThreadTools ?? []).map((tool) => ({
                    tool,
                    variant: ToolVariant.AVAILABLE,
                })),
            ],
        };
    }, [agent]);

    const form = useForm<ToolFormValues>({
        resolver: zodResolver(formSchema),
        defaultValues,
    });

    const toolFields = useFieldArray({
        control: form.control,
        name: "tools",
    });

    const handleSubmit = form.handleSubmit(onSubmit || noop);

    useEffect(() => {
        return form.watch((values) => {
            const { data, success } = formSchema.safeParse(values);

            if (!success) return;

            onChange?.(data);
        }).unsubscribe;
    }, [form, onChange]);

    const [fixedFields, userFields] = useMemo(() => {
        return [
            toolFields.fields?.filter(
                (field) => field.variant === ToolVariant.FIXED
            ),
            toolFields.fields?.filter(
                (field) => field.variant !== ToolVariant.FIXED
            ),
        ];
    }, [toolFields]);

    const removeTool = (tool: string) =>
        toolFields.remove(toolFields.fields.findIndex((t) => t.tool === tool));

    const addTool = (tool: string) =>
        toolFields.append({ tool, variant: ToolVariant.FIXED });

    const updateVariant = (tool: string, variant: ToolVariant) =>
        toolFields.update(
            toolFields.fields.findIndex((t) => t.tool === tool),
            { tool, variant }
        );

    return (
        <Form {...form}>
            <form onSubmit={handleSubmit} className="flex flex-col gap-2">
                <TypographyH4 className="flex justify-between items-end">
                    <span className="min-w-fit">Agent Tools</span>
                    {renderAddButton()}
                </TypographyH4>

                <TypographySmall>
                    These tools are essential for the agent&apos;s core
                    functionality and are always enabled.
                </TypographySmall>

                <div className="mt-2 w-full overflow-y-auto">
                    {fixedFields.map((field) => (
                        <ToolEntry
                            key={field.id}
                            tool={field.tool}
                            onDelete={() => removeTool(field.tool)}
                            actions={renderFixedActions(field.tool)}
                        />
                    ))}
                </div>

                <TypographyH4 className="mt-4">User Tools</TypographyH4>

                <TypographySmall>
                    Optional tools users can turn on or off. Use the toggle to
                    switch whether they&apos;re active by default by the agent.
                </TypographySmall>

                <div className="mt-2 w-full overflow-y-auto">
                    {userFields.map((field) => (
                        <ToolEntry
                            key={field.id}
                            tool={field.tool}
                            onDelete={() => removeTool(field.tool)}
                            actions={renderDefaultActions(
                                field.tool,
                                field.variant
                            )}
                        />
                    ))}
                </div>
            </form>
        </Form>
    );

    function renderFixedActions(tool: string) {
        return (
            <TooltipProvider>
                <Tooltip>
                    <TooltipTrigger asChild>
                        <Button
                            variant="secondary"
                            size="icon"
                            onClick={() =>
                                updateVariant(tool, ToolVariant.DEFAULT)
                            }
                        >
                            <ArrowDownIcon className="w-4 h-4" />
                        </Button>
                    </TooltipTrigger>

                    <TooltipContent>
                        Make this tool optional for users
                    </TooltipContent>
                </Tooltip>
            </TooltipProvider>
        );
    }

    function renderDefaultActions(tool: string, variant: ToolVariant) {
        return (
            <>
                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger
                            className="flex items-center gap-2"
                            asChild
                        >
                            <div>
                                <FormLabel htmlFor="default-switch">
                                    Default
                                </FormLabel>

                                <Switch
                                    checked={variant === ToolVariant.DEFAULT}
                                    name="default-switch"
                                    onCheckedChange={(checked) =>
                                        updateVariant(
                                            tool,
                                            checked
                                                ? ToolVariant.DEFAULT
                                                : ToolVariant.AVAILABLE
                                        )
                                    }
                                />
                            </div>
                        </TooltipTrigger>

                        <TooltipContent>
                            {variant === ToolVariant.DEFAULT
                                ? "This tool is available by default"
                                : "This tool can be added by a user"}
                        </TooltipContent>
                    </Tooltip>
                </TooltipProvider>

                <TooltipProvider>
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="secondary"
                                size="icon"
                                onClick={() =>
                                    updateVariant(tool, ToolVariant.FIXED)
                                }
                            >
                                <ArrowUpIcon className="w-4 h-4" />
                            </Button>
                        </TooltipTrigger>

                        <TooltipContent>
                            Make this tool essential for the agent
                        </TooltipContent>
                    </Tooltip>
                </TooltipProvider>
            </>
        );
    }

    function renderAddButton() {
        return (
            <div className="flex justify-end w-full">
                <Dialog>
                    <DialogTrigger asChild>
                        <Button variant="secondary">
                            <PlusIcon className="w-4 h-4 mr-2" /> Add Tool
                        </Button>
                    </DialogTrigger>

                    <DialogContent className="p-0 max-w-3xl min-h-[350px]">
                        <DialogTitle hidden>Tool Catalog</DialogTitle>
                        <DialogDescription hidden>
                            Add tools to the agent.
                        </DialogDescription>

                        <ToolCatalog
                            className="w-full border-none"
                            tools={toolFields.fields.map((field) => field.tool)}
                            onAddTool={addTool}
                            onRemoveTool={removeTool}
                        />
                    </DialogContent>
                </Dialog>
            </div>
        );
    }
}
