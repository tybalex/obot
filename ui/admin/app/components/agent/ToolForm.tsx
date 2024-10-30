import { zodResolver } from "@hookform/resolvers/zod";
import { PlusIcon } from "lucide-react";
import { useCallback, useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Agent } from "~/lib/model/agents";
import { noop } from "~/lib/utils";

import { ToolEntry } from "~/components/agent/ToolEntry";
import { ToolCatalog } from "~/components/tools/ToolCatalog";
import { Button } from "~/components/ui/button";
import { Dialog, DialogContent, DialogTrigger } from "~/components/ui/dialog";
import {
    Form,
    FormDescription,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "~/components/ui/form";

const formSchema = z.object({
    tools: z.array(z.string()),
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
    const form = useForm<ToolFormValues>({
        resolver: zodResolver(formSchema),
        defaultValues: { tools: agent.tools || [] },
    });

    const handleSubmit = form.handleSubmit(onSubmit || noop);

    const toolValues = form.watch("tools");

    useEffect(() => {
        return form.watch((values) => {
            const { data, success } = formSchema.safeParse(values);

            if (!success) return;

            onChange?.(data);
        }).unsubscribe;
    }, [toolValues, form.formState, onChange, form]);

    const handleToolsChange = useCallback(
        (newTools: string[]) => {
            form.setValue("tools", newTools, {
                shouldValidate: true,
                shouldDirty: true,
            });
            onChange?.({ tools: newTools });
        },
        [form, onChange]
    );

    return (
        <Form {...form}>
            <form onSubmit={handleSubmit} className="flex flex-col gap-2">
                <FormField
                    control={form.control}
                    name="tools"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel hidden>Tools</FormLabel>
                            <FormDescription hidden>
                                The tools of the agent.
                            </FormDescription>

                            <div className="mt-2 w-full overflow-y-auto">
                                {field.value?.map((tool, index) => (
                                    <ToolEntry
                                        key={tool}
                                        tool={tool}
                                        onDelete={() => {
                                            const newTools =
                                                field.value?.filter(
                                                    (_, i) => i !== index
                                                );

                                            field.onChange(newTools);
                                        }}
                                    />
                                ))}
                            </div>
                            <div className="flex justify-end w-full my-4">
                                <Dialog>
                                    <DialogTrigger asChild>
                                        <Button
                                            variant="secondary"
                                            className="mt-4 mb-4"
                                        >
                                            <PlusIcon className="w-4 h-4 mr-2" />{" "}
                                            Add Tool
                                        </Button>
                                    </DialogTrigger>
                                    <DialogContent className="p-0 max-w-3xl min-h-[350px]">
                                        <ToolCatalog
                                            className="w-full border-none"
                                            tools={toolValues}
                                            onChangeTools={handleToolsChange}
                                        />
                                    </DialogContent>
                                </Dialog>
                            </div>
                            <FormMessage />
                        </FormItem>
                    )}
                />
            </form>
        </Form>
    );
}
