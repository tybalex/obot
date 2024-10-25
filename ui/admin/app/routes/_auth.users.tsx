import { Link } from "@remix-run/react";
import { ColumnDef, createColumnHelper } from "@tanstack/react-table";
import useSWR from "swr";

import { User, roleToString } from "~/lib/model/users";
import { ThreadsService } from "~/lib/service/api/threadsService";
import { UserService } from "~/lib/service/api/userService";
import { timeSince } from "~/lib/utils";

import { TypographyH2, TypographyP } from "~/components/Typography";
import { DataTable } from "~/components/composed/DataTable";

export default function Users() {
    const getUsers = useSWR(UserService.getUsers.key(), UserService.getUsers);
    const users = getUsers.data || [];

    const { data: threads } = useSWR(
        () => users.length > 0 && ThreadsService.getThreads.key(),
        () => ThreadsService.getThreads()
    );

    const threadIdSet = new Set(threads?.map((thread) => thread.id) || []);

    return (
        <div>
            <div className="h-full p-8 flex flex-col gap-4">
                <TypographyH2 className="mb-4">Users</TypographyH2>
                <DataTable
                    columns={getColumns(threadIdSet)}
                    data={users}
                    sort={[{ id: "created", desc: true }]}
                />
            </div>
        </div>
    );

    function getColumns(threadIdSet: Set<string>): ColumnDef<User, string>[] {
        return [
            columnHelper.accessor("email", {
                header: "Email",
            }),
            columnHelper.display({
                id: "thread",
                header: "Thread",
                cell: ({ row }) => {
                    const threadId = `t1${row.original.id}`;
                    if (threadIdSet.has(threadId)) {
                        return (
                            <Link
                                to={`/thread/${threadId}?from=/users`}
                                className="underline"
                            >
                                View Thread
                            </Link>
                        );
                    }
                    return <TypographyP>No Threads</TypographyP>;
                },
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
