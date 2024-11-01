import { TrashIcon } from "lucide-react";
import useSWR from "swr";

import { ToolReferenceService } from "~/lib/service/api/toolreferenceService";

import { ToolIcon } from "~/components/tools/ToolIcon";
import { LoadingSpinner } from "~/components/ui/LoadingSpinner";
import { Button } from "~/components/ui/button";

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
            <div className="border text-sm px-3 shadow-sm rounded-md p-2 w-full flex items-center gap-2">
                {isLoading ? (
                    <LoadingSpinner className="w-4 h-4" />
                ) : (
                    <ToolIcon
                        className="w-4 h-4"
                        name={toolReference?.name || tool}
                        icon={toolReference?.metadata?.icon}
                    />
                )}
                {toolReference?.name || tool}
            </div>
            <Button
                type="button"
                variant="secondary"
                size="icon"
                onClick={() => onDelete()}
            >
                <TrashIcon className="w-4 h-4" />
            </Button>
        </div>
    );
}
