import { zodResolver } from "@hookform/resolvers/zod";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { Agent } from "~/lib/model/agents";

import { ControlledTextarea } from "~/components/form/controlledInputs";
import { Form } from "~/components/ui/form";

const formSchema = z.object({
    prompt: z.string().optional(),
});

export type AdvancedFormValues = z.infer<typeof formSchema>;

type AdvancedFormProps = {
    agent: Agent;
    onSubmit?: (values: AdvancedFormValues) => void;
    onChange?: (values: AdvancedFormValues) => void;
};

export function AdvancedForm({ agent, onSubmit, onChange }: AdvancedFormProps) {
    const form = useForm<AdvancedFormValues>({
        resolver: zodResolver(formSchema),
        mode: "onChange",
        defaultValues: {
            prompt: agent.prompt || "",
        },
    });

    useEffect(() => {
        if (agent) form.reset({ prompt: agent.prompt || "" });
    }, [agent, form]);

    useEffect(() => {
        return form.watch((values) => {
            if (!onChange) return;

            const { data, success } = formSchema.safeParse(values);

            if (!success) return;

            onChange(data);
        }).unsubscribe;
    }, [onChange, form]);

    const handleSubmit = form.handleSubmit((values: AdvancedFormValues) =>
        onSubmit?.({ ...values })
    );

    return (
        <Form {...form}>
            <form onSubmit={handleSubmit} className="space-y-8">
                <ControlledTextarea
                    control={form.control}
                    name="prompt"
                    label="Additional Instructions"
                    description="Give the agent additional instructions on how it should behave."
                    placeholder="Talk like a pirate."
                />
            </form>
        </Form>
    );
}
