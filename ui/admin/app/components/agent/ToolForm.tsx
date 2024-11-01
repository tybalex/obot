import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon } from "lucide-react";
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
import { Form } from "~/components/ui/form";

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

const getVariant = (tool: string, agent: Agent): ToolVariant => {
    if (agent.defaultThreadTools?.includes(tool)) return "default";
    if (agent.availableThreadTools?.includes(tool)) return "available";
    return "fixed";
};

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
            tools: agent.tools?.map((tool) => ({
                tool,
                variant: getVariant(tool, agent),
            })),
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

            console.log(data);

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

    return (
        <Form {...form}>
            <form onSubmit={handleSubmit} className="flex flex-col gap-2">
                <div className="mt-2 w-full overflow-y-auto">
                    {fixedFields.map((field) => (
                        <ToolEntry
                            key={field.id}
                            tool={field.tool}
                            onDelete={() => removeTool(field.tool)}
                        />
                    ))}
                </div>

                <div className="flex justify-end w-full my-4">
                    <Dialog>
                        <DialogTrigger asChild>
                            <Button variant="secondary" className="mt-4 mb-4">
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
                                tools={toolFields.fields.map(
                                    (field) => field.tool
                                )}
                                onAddTool={addTool}
                                onRemoveTool={removeTool}
                            />
                        </DialogContent>
                    </Dialog>
                </div>
            </form>
        </Form>
    );
}
