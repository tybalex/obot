import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useMemo } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { z } from "zod";

import { Agent } from "~/lib/model/agents";
import { noop } from "~/lib/utils";

import { ToolEntry } from "~/components/agent/ToolEntry";
import { ToolCatalogDialog } from "~/components/tools/ToolCatalog";
import { Form } from "~/components/ui/form";
import { Switch } from "~/components/ui/switch";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "~/components/ui/tooltip";

import { TypographyP, TypographySmall } from "../Typography";

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

    const updateVariant = (tool: string, variant: ToolVariant) =>
        toolFields.update(
            toolFields.fields.findIndex((t) => t.tool === tool),
            { tool, variant }
        );

    const allTools = useMemo(() => {
        return toolFields.fields.map(({ tool }) => tool);
    }, [toolFields]);

    return (
        <Form {...form}>
            <form onSubmit={handleSubmit} className="flex flex-col gap-2">
                <TypographyP className="flex justify-between items-end font-normal">
                    Agent Tools
                </TypographyP>

                <TypographySmall className="text-muted-foreground">
                    These tools are essential for the agent&apos;s core
                    functionality and are always enabled.
                </TypographySmall>

                <div className="mt-2 w-full overflow-y-auto">
                    {fixedFields.map((field) => (
                        <ToolEntry
                            key={field.id}
                            tool={field.tool}
                            onDelete={() => removeTool(field.tool)}
                        />
                    ))}
                </div>

                <ToolCatalogDialog
                    tools={allTools}
                    onAddTool={(tool) =>
                        toolFields.append({
                            tool,
                            variant: ToolVariant.FIXED,
                        })
                    }
                    onRemoveTool={removeTool}
                />

                <TypographyP className="flex justify-between items-end font-normal mt-4">
                    User Tools
                </TypographyP>

                <TypographySmall className="text-muted-foreground">
                    Optional tools users can turn on or off. Use the toggle to
                    set whether they&apos;re active by default for the agent.
                </TypographySmall>

                <div className="mt-2 w-full overflow-y-auto">
                    {userFields.map((field) => (
                        <ToolEntry
                            key={field.id}
                            tool={field.tool}
                            onDelete={() => removeTool(field.tool)}
                            actions={
                                <TooltipProvider>
                                    <Tooltip>
                                        <TooltipTrigger>
                                            <Switch
                                                checked={
                                                    field.variant ===
                                                    ToolVariant.DEFAULT
                                                }
                                                onCheckedChange={(checked) =>
                                                    updateVariant(
                                                        field.tool,
                                                        checked
                                                            ? ToolVariant.DEFAULT
                                                            : ToolVariant.AVAILABLE
                                                    )
                                                }
                                            />
                                        </TooltipTrigger>

                                        <TooltipContent>
                                            {field.variant ===
                                            ToolVariant.DEFAULT
                                                ? "Active by Default"
                                                : "Inactive by Default"}
                                        </TooltipContent>
                                    </Tooltip>
                                </TooltipProvider>
                            }
                        />
                    ))}
                </div>

                <ToolCatalogDialog
                    tools={allTools}
                    onAddTool={(tool) =>
                        toolFields.append({
                            tool,
                            variant: ToolVariant.DEFAULT,
                        })
                    }
                    onRemoveTool={removeTool}
                />
            </form>
        </Form>
    );
}
