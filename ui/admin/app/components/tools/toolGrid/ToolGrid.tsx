import { useMemo } from "react";

import { ToolReference } from "~/lib/model/toolReferences";
import { ToolCategoryMap } from "~/lib/service/api/toolreferenceService";

import { CategoryHeader } from "~/components/tools/toolGrid/CategoryHeader";
import { CategoryTools } from "~/components/tools/toolGrid/CategoryTools";

interface ToolGridProps {
    toolCategories: ToolCategoryMap;
    filter: string;
    onDelete: (id: string) => void;
}

export function ToolGrid({ toolCategories, filter, onDelete }: ToolGridProps) {
    const filteredCategories = useMemo(() => {
        const result: ToolCategoryMap = {};
        for (const [category, { tools, bundleTool }] of Object.entries(
            toolCategories
        )) {
            const filteredTools = tools.filter(
                (tool) =>
                    tool.name?.toLowerCase().includes(filter.toLowerCase()) ||
                    tool.metadata?.category
                        ?.toLowerCase()
                        .includes(filter.toLowerCase()) ||
                    tool.description
                        ?.toLowerCase()
                        .includes(filter.toLowerCase())
            );
            if (filteredTools.length > 0 || bundleTool) {
                result[category] = {
                    tools: filteredTools,
                    bundleTool: bundleTool,
                };
            }
        }
        return result;
    }, [toolCategories, filter]);

    return (
        <div className="space-y-8 pb-16">
            {Object.entries(filteredCategories).map(
                ([category, { tools, bundleTool }]) => (
                    <div key={category} className="space-y-4">
                        <CategoryHeader category={category} tools={tools} />
                        <CategoryTools
                            tools={[
                                bundleTool || ({} as ToolReference),
                                ...tools,
                            ]}
                            onDelete={onDelete}
                        />
                    </div>
                )
            )}
        </div>
    );
}
