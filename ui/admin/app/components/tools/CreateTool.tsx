import { useForm } from "react-hook-form";

import { CreateToolReference } from "~/lib/model/toolReferences";
import { ToolReferenceService } from "~/lib/service/api/toolreferenceService";

import { Button } from "~/components/ui/button";
import { Input } from "~/components/ui/input";
import { useAsync } from "~/hooks/useAsync";

interface CreateToolProps {
    onSuccess: () => void;
}

export function CreateTool({ onSuccess }: CreateToolProps) {
    const { register, handleSubmit, reset } = useForm<CreateToolReference>();

    const { execute: onSubmit, isLoading } = useAsync(
        async (data: CreateToolReference) => {
            await ToolReferenceService.createToolReference({
                toolReference: { ...data, toolType: "tool" },
            });
            reset();
            onSuccess();
        },
        {
            onError: (error) =>
                console.error("Failed to create tool reference:", error),
        }
    );

    return (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
            <div>
                <Input
                    autoComplete="off"
                    {...register("name", { required: "Name is required" })}
                    placeholder="Tool Name"
                />
            </div>
            <div>
                <Input
                    autoComplete="off"
                    {...register("reference", {
                        required: "Reference is required",
                    })}
                    placeholder="Tool Reference"
                />
            </div>
            <div>
                <Button type="submit" variant="secondary" disabled={isLoading}>
                    {isLoading ? "Creating..." : "Register Tool"}
                </Button>
            </div>
        </form>
    );
}
