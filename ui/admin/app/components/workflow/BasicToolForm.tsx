import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ToolEntry } from "~/components/agent/ToolEntry";
import { ToolCatalogDialog } from "~/components/tools/ToolCatalog";
import { Form } from "~/components/ui/form";

const formSchema = z.object({
    tools: z.array(z.string()),
});

export type BasicToolFormValues = z.infer<typeof formSchema>;

export function BasicToolForm({
    defaultValues,
    onChange,
}: {
    defaultValues?: Partial<BasicToolFormValues>;
    onSubmit?: (values: BasicToolFormValues) => void;
    onChange?: (values: BasicToolFormValues) => void;
}) {
    const form = useForm<BasicToolFormValues>({
        resolver: zodResolver(formSchema),
        defaultValues,
    });

    useEffect(() => {
        return form.watch((values) => {
            const { data, success } = formSchema.safeParse(values);

            if (!success) return;

            onChange?.(data);
        }).unsubscribe;
    }, [form, onChange]);

    const tools = form.watch("tools");

    const removeTools = (toolsToRemove: string[]) => {
        form.setValue(
            "tools",
            tools.filter((t) => !toolsToRemove.includes(t))
        );
    };

    const addTool = (tool: string) => form.setValue("tools", [...tools, tool]);

    return (
        <Form {...form}>
            <div className="flex flex-col gap-2">
                <div className="flex flex-col gap-1 w-full overflow-y-auto">
                    {tools.map((tool) => (
                        <ToolEntry
                            key={tool}
                            tool={tool}
                            onDelete={() => removeTools([tool])}
                        />
                    ))}
                </div>

                <div className="flex justify-end">
                    <ToolCatalogDialog
                        tools={tools}
                        onAddTool={addTool}
                        onRemoveTools={removeTools}
                    />
                </div>
            </div>
        </Form>
    );
}
