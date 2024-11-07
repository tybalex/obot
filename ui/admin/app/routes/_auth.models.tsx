import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import { useMemo } from "react";
import useSWR, { preload } from "swr";

import { Model } from "~/lib/model/models";
import { ModelApiService } from "~/lib/service/api/modelApiService";

import { TypographyH2, TypographySmall } from "~/components/Typography";
import { DataTable } from "~/components/composed/DataTable";
import { CreateModel } from "~/components/model/CreateModel";
import { DeleteModel } from "~/components/model/DeleteModel";
import { UpdateModel } from "~/components/model/UpdateModel";

export async function clientLoader() {
    await Promise.all([
        preload(ModelApiService.getModels.key(), ModelApiService.getModels),
        preload(
            ModelApiService.getModelProviders.key(true),
            ({ onlyConfigured }) =>
                ModelApiService.getModelProviders(onlyConfigured)
        ),
    ]);
    return null;
}

export default function Models() {
    const { data } = useSWR(
        ModelApiService.getModels.key(),
        ModelApiService.getModels
    );

    const { data: providers } = useSWR(
        ModelApiService.getModelProviders.key(true),
        ({ onlyConfigured }) =>
            ModelApiService.getModelProviders(onlyConfigured)
    );

    const providerMap = useMemo(() => {
        if (!providers) return {};
        return providers?.reduce(
            (acc, provider) => {
                acc[provider.id] = provider.name;
                return acc;
            },
            {} as Record<string, string>
        );
    }, [providers]);

    return (
        <div className="h-full flex flex-col p-8 space-y-4">
            <div className="flex items-center justify-between">
                <TypographyH2>Models</TypographyH2>
                <CreateModel />
            </div>

            <DataTable
                columns={getColumns()}
                data={data ?? []}
                sort={[{ id: "id", desc: true }]}
                classNames={{
                    row: "!max-h-[200px] grow-0 height-[200px]",
                    cell: "!max-h-[200px] grow-0 height-[200px]",
                }}
                disableClickPropagation={(cell) => cell.id.includes("actions")}
            />
        </div>
    );

    function getColumns(): ColumnDef<Model, string>[] {
        return [
            columnHelper.accessor((model) => model.name ?? model.id, {
                id: "id",
                header: "Model",
            }),
            columnHelper.accessor(
                (model) =>
                    providerMap[model.modelProvider] ?? model.modelProvider,
                {
                    id: "provider",
                    header: "Provider",
                }
            ),
            columnHelper.display({
                id: "default",
                cell: ({ row }) => {
                    const value = row.original.default;

                    if (!value) return null;

                    return (
                        <TypographySmall className="text-muted-foreground flex items-center gap-2">
                            <div className="size-2 bg-success rounded-full" />
                            Default
                        </TypographySmall>
                    );
                },
            }),
            columnHelper.display({
                id: "actions",
                cell: ({ row }) => (
                    <div className="flex justify-end">
                        <UpdateModel model={row.original} />
                        <DeleteModel id={row.original.id} />
                    </div>
                ),
            }),
        ];
    }
}

const columnHelper = createColumnHelper<Model>();
