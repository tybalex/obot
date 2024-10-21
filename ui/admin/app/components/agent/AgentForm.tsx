import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Agent } from "~/lib/model/agents";

import { Form } from "~/components/ui/form";

import { ControlledInput } from "../form/controlledInputs";

const formSchema = z.object({
    name: z.string().min(1, {
        message: "Name is required.",
    }),
    description: z.string().optional(),
    prompt: z.string().optional(),
});

export type AgentInfoFormValues = z.infer<typeof formSchema>;

type AgentFormProps = {
    agent: Agent;
    onSubmit?: (values: AgentInfoFormValues) => void;
    onChange?: (values: AgentInfoFormValues) => void;
};

export function AgentForm({ agent, onSubmit, onChange }: AgentFormProps) {
    const form = useForm<AgentInfoFormValues>({
        resolver: zodResolver(formSchema),
        mode: "onChange",
        defaultValues: {
            name: agent.name || "",
            description: agent.description || "",
            prompt: agent.prompt || "",
        },
    });

    useEffect(() => {
        if (agent) form.reset(agent);
    }, [agent, form]);

    useEffect(() => {
        return form.watch((values) => {
            if (!onChange) return;

            const { data, success } = formSchema.safeParse(values);

            if (!success) return;

            onChange(data);
        }).unsubscribe;
    }, [onChange, form]);

    const handleSubmit = form.handleSubmit((values: AgentInfoFormValues) =>
        onSubmit?.({ ...agent, ...values })
    );

    return (
        <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-4">
                <ControlledInput
                    autoComplete="off"
                    control={form.control}
                    name="name"
                    className="text-3xl shadow-none cursor-pointer hover:border-primary px-0 mb-0 font-bold outline-none border-transparent focus:border-primary"
                />
                <ControlledInput
                    control={form.control}
                    autoComplete="off"
                    name="description"
                    placeholder="Add a description..."
                    className="text-xl text-muted-foreground font-semibold shadow-none cursor-pointer hover:border-primary px-0 outline-none border-transparent focus:border-primary"
                />
            </form>
        </Form>
    );
}
