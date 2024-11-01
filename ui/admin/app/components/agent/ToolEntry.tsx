import { zodResolver } from "@hookform/resolvers/zod";
import { TrashIcon } from "lucide-react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import useSWR from "swr";
import { z } from "zod";

import { ToolReferenceService } from "~/lib/service/api/toolreferenceService";
import { noop } from "~/lib/utils";

import { TruncatedText } from "../TruncatedText";
import { ToolIcon } from "../tools/ToolIcon";
import { LoadingSpinner } from "../ui/LoadingSpinner";
import { Button } from "../ui/button";
import { Form, FormField, FormItem, FormMessage } from "../ui/form";
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "../ui/select";

export function ToolEntry({
    tool,
    onDelete,
}: {
    tool: string;
    onDelete: () => void;
}) {
    const { data: toolReference, isLoading } = useSWR(
        ToolReferenceService.getToolReferenceById.key(tool),
        ({ toolReferenceId }) =>
            ToolReferenceService.getToolReferenceById(toolReferenceId)
    );

    return (
        <div className="flex items-center space-x-2 justify-between mt-1">
            <div className="border text-sm px-3 shadow-sm rounded-md p-2 w-full flex items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                    {isLoading ? (
                        <LoadingSpinner className="w-5 h-5" />
                    ) : (
                        <ToolIcon
                            className="w-5 h-5"
                            name={toolReference?.name || tool}
                            icon={toolReference?.metadata?.icon}
                        />
                    )}

                    <TruncatedText content={toolReference?.name || tool} />
                </div>

                <div className="flex items-center gap-2">
                    <ToolEntryForm onChange={noop} />

                    <Button
                        type="button"
                        variant="secondary"
                        size="icon"
                        onClick={() => onDelete()}
                    >
                        <TrashIcon className="w-5 h-5" />
                    </Button>
                </div>
            </div>
        </div>
    );
}

const schema = z.object({
    variant: z.enum(["fixed", "default", "canAdd"]),
});

type ToolEntryForm = z.infer<typeof schema>;

function ToolEntryForm({
    onChange,
}: {
    onChange: (data: ToolEntryForm) => void;
}) {
    const form = useForm<ToolEntryForm>({
        resolver: zodResolver(schema),
        defaultValues: {
            variant: "default",
        },
    });

    useEffect(() => {
        return form.watch((values) => {
            const { success, data } = schema.safeParse(values);

            if (!success) return;

            onChange(data);
        }).unsubscribe;
    }, [form, onChange]);

    return (
        <Form {...form}>
            <FormField
                control={form.control}
                name="variant"
                render={({ field: { ref: _, ...field } }) => (
                    <FormItem>
                        <Select {...field} onValueChange={field.onChange}>
                            <SelectTrigger>
                                <SelectValue />
                            </SelectTrigger>

                            <SelectContent>
                                <SelectItem value="fixed">Fixed</SelectItem>
                                <SelectItem value="default">Default</SelectItem>
                                <SelectItem value="canAdd">Can Add</SelectItem>
                            </SelectContent>
                        </Select>

                        <FormMessage />
                    </FormItem>
                )}
            />
        </Form>
    );
}
