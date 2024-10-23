import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { Trash } from "lucide-react";

import { ToolReference } from "~/lib/model/toolReferences";
import { timeSince } from "~/lib/utils";

import { TypographyP } from "~/components/Typography";
import { ConfirmationDialog } from "~/components/composed/ConfirmationDialog";
import { DataTable } from "~/components/composed/DataTable";
import { ToolIcon } from "~/components/tools/ToolIcon";
import { Button } from "~/components/ui/button";

interface ToolTableProps {
    tools: ToolReference[];
    filter: string;
    onDelete: (id: string) => void;
}

export function ToolTable({ tools, filter, onDelete }: ToolTableProps) {
    const filteredTools = tools.filter(
        (tool) =>
            tool.name?.toLowerCase().includes(filter.toLowerCase()) ||
            tool.metadata?.category
                ?.toLowerCase()
                .includes(filter.toLowerCase()) ||
            tool.description?.toLowerCase().includes(filter.toLowerCase())
    );

    return (
        <DataTable
            columns={getColumns(onDelete)}
            data={filteredTools}
            sort={[{ id: "created", desc: true }]}
        />
    );
}

function getColumns(
    onDelete: (id: string) => void
): ColumnDef<ToolReference, string>[] {
    const columnHelper = createColumnHelper<ToolReference>();

    return [
        columnHelper.display({
            id: "category",
            header: "Category",
            cell: ({ row }) => (
                <TypographyP className="flex items-center gap-2">
                    <ToolIcon
                        className="w-5 h-5"
                        name={row.original.name}
                        icon={row.original.metadata?.icon}
                    />
                    {row.original.metadata?.category ?? "Uncategorized"}
                </TypographyP>
            ),
        }),
        columnHelper.display({
            id: "name",
            header: "Name",
            cell: ({ row }) => (
                <TypographyP>
                    {row.original.name}
                    {row.original.metadata?.bundle ? " Bundle" : ""}
                </TypographyP>
            ),
        }),
        columnHelper.accessor("reference", {
            header: "Reference",
        }),
        columnHelper.display({
            id: "description",
            header: "Description",
            cell: ({ row }) => (
                <TypographyP>{row.original.description}</TypographyP>
            ),
        }),
        columnHelper.accessor("created", {
            header: "Created",
            cell: ({ getValue }) => (
                <TypographyP>{timeSince(new Date(getValue()))} ago</TypographyP>
            ),
            sortingFn: "datetime",
        }),
        columnHelper.display({
            id: "actions",
            cell: ({ row }) => (
                <ConfirmationDialog
                    title="Delete Tool Reference"
                    description="Are you sure you want to delete this tool reference? This action cannot be undone."
                    onConfirm={() => onDelete(row.original.id)}
                    confirmProps={{
                        variant: "destructive",
                        children: "Delete",
                    }}
                >
                    <Button variant="ghost" size="sm">
                        <Trash className="h-4 w-4" />
                    </Button>
                </ConfirmationDialog>
            ),
        }),
    ];
}
