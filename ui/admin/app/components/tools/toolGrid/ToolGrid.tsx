import { useMemo } from "react";

import { ToolReference } from "~/lib/model/toolReferences";

import { CategoryHeader } from "~/components/tools/toolGrid/CategoryHeader";
import { CategoryTools } from "~/components/tools/toolGrid/CategoryTools";

interface ToolGridProps {
    tools: ToolReference[];
    filter: string;
    onDelete: (id: string) => void;
}

export function ToolGrid({ tools, filter, onDelete }: ToolGridProps) {
    const filteredTools = useMemo(() => {
        return tools?.filter(
            (tool) =>
                tool.name?.toLowerCase().includes(filter.toLowerCase()) ||
                tool.metadata?.category
                    ?.toLowerCase()
                    .includes(filter.toLowerCase()) ||
                tool.description?.toLowerCase().includes(filter.toLowerCase())
        );
    }, [tools, filter]);

    const sortedCategories = useMemo(() => {
        const categorizedTools = filteredTools?.reduce(
            (acc, tool) => {
                const category = tool.metadata?.category ?? "Uncategorized";
                if (!acc[category]) {
                    acc[category] = [];
                }
                acc[category].push(tool);
                return acc;
            },
            {} as Record<string, ToolReference[]>
        );

        // Sort categories to put "Uncategorized" first
        return Object.entries(categorizedTools).sort(([a], [b]) => {
            if (a === "Uncategorized") return -1;
            if (b === "Uncategorized") return 1;
            return a.localeCompare(b);
        });
    }, [filteredTools]);

    return (
        <div className="space-y-8 pb-16">
            {sortedCategories.map(([category, categoryTools]) => (
                <div key={category} className="space-y-4">
                    <CategoryHeader category={category} tools={categoryTools} />
                    <CategoryTools tools={categoryTools} onDelete={onDelete} />
                </div>
            ))}
        </div>
    );
}
