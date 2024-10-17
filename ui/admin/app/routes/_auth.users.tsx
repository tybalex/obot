import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import useSWR from "swr";

import { User, roleToString } from "~/lib/model/users";
import { UserService } from "~/lib/service/api/userService";
import { timeSince } from "~/lib/utils";

import { TypographyP } from "~/components/Typography";
import { DataTable } from "~/components/composed/DataTable";

export default function Users() {
    const getUsers = useSWR(UserService.getUsers.key(), UserService.getUsers);

    const users = getUsers.data || [];

    return (
        <div>
            <div className="h-full p-8 flex flex-col gap-4">
                <DataTable
                    columns={getColumns()}
                    data={users}
                    sort={[{ id: "created", desc: true }]}
                />
            </div>
        </div>
    );

    function getColumns(): ColumnDef<User, string>[] {
        return [
            columnHelper.accessor("email", {
                header: "Email",
            }),
            columnHelper.accessor("username", {
                header: "Username",
            }),
            columnHelper.display({
                id: "role",
                header: "Role",
                cell: ({ row }) => (
                    <TypographyP>{roleToString(row.original.role)}</TypographyP>
                ),
            }),
            columnHelper.display({
                id: "created",
                header: "Created",
                cell: ({ row }) => (
                    <TypographyP>
                        {timeSince(new Date(row.original.created))} ago
                    </TypographyP>
                ),
            }),
        ];
    }
}

const columnHelper = createColumnHelper<User>();
