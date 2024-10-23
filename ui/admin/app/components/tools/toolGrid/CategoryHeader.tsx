import { Folder } from "lucide-react";

import { ToolReference } from "~/lib/model/toolReferences";

import { TypographyH2 } from "~/components/Typography";
import { ToolIcon } from "~/components/tools/ToolIcon";
import { Badge } from "~/components/ui/badge";

interface CategoryHeaderProps {
    category: string;
    tools: ToolReference[];
}

export function CategoryHeader({ category, tools }: CategoryHeaderProps) {
    return (
        <div className="flex items-center space-x-4">
            <div className="w-10 h-10 flex items-center justify-center bg-muted rounded-full mb-2 border">
                {tools[0]?.metadata?.icon ? (
                    <ToolIcon
                        className="w-6 h-6"
                        name={tools[0].name}
                        icon={tools[0].metadata.icon}
                    />
                ) : (
                    <Folder className="w-6 h-6" />
                )}
            </div>
            <TypographyH2 className="flex items-center space-x-2">
                <span>{category}</span>
                <Badge>{tools.length}</Badge>
            </TypographyH2>
        </div>
    );
}
