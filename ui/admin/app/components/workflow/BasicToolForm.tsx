import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect, useMemo, useRef } from "react";
import { useFieldArray, useForm } from "react-hook-form";
import { z } from "zod";

import { ToolEntry } from "~/components/agent/ToolEntry";
import { ToolCatalogDialog } from "~/components/tools/ToolCatalog";
import { Form } from "~/components/ui/form";

const formSchema = z.object({
    tools: z.array(z.object({ value: z.string() })),
});

export type BasicToolFormValues = z.infer<typeof formSchema>;

export function BasicToolForm({
    defaultValues: _defaultValues,
    onChange,
}: {
    defaultValues?: { tools?: string[] };
    onSubmit?: (values: BasicToolFormValues) => void;
    onChange?: (values: BasicToolFormValues) => void;
}) {
    const defaultValues = useMemo(() => {
        return {
            tools:
                _defaultValues?.tools?.map((tool) => ({ value: tool })) || [],
        };
    }, [_defaultValues]);

    const form = useForm<BasicToolFormValues>({
        resolver: zodResolver(formSchema),
        defaultValues,
    });

    const dirty = useRef(false);
    dirty.current = form.formState.isDirty;

    useEffect(() => {
        return form.watch((values) => {
            if (!dirty.current) return;

            const { data, success } = formSchema.safeParse(values);

            if (!success) return;

            onChange?.(data);
        }).unsubscribe;
    }, [form, onChange]);

    const fields = useFieldArray({ control: form.control, name: "tools" });

    const currentTools = useMemo(
        () => fields.fields.map((field) => field.value),
        [fields.fields]
    );

    const removeTools = (toolsToRemove: string[]) => {
        const indexes = toolsToRemove
            .map((tool) => fields.fields.findIndex((t) => t.value === tool))
            .filter((index) => index !== -1);

        indexes.forEach((index) => fields.remove(index));
    };

    return (
        <Form {...form}>
            <div className="flex flex-col gap-2">
                <div className="flex flex-col gap-1 w-full overflow-y-auto">
                    {fields.fields.map((field) => (
                        <ToolEntry
                            key={field.id}
                            tool={field.value}
                            onDelete={() => removeTools([field.value])}
                        />
                    ))}
                </div>

                <div className="flex justify-end">
                    <ToolCatalogDialog
                        tools={currentTools}
                        onAddTool={(tool) => fields.append({ value: tool })}
                        onRemoveTools={removeTools}
                    />
                </div>
            </div>
        </Form>
    );
}
