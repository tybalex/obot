import { Outlet, redirect } from "@remix-run/react";
import { isAxiosError } from "axios";
import { $path } from "remix-routes";

import { UserService } from "~/lib/service/api/userService";

import { HeaderNav } from "~/components/header/HeaderNav";
import { Sidebar } from "~/components/sidebar";

export const clientLoader = async () => {
    try {
        await UserService.getMe();
    } catch (error) {
        if (isAxiosError(error) && error.response?.status === 403) {
            throw redirect($path("/sign-in"));
        }
    }
    return null;
};

export default function AuthLayout() {
    return (
        <div className="flex h-screen w-screen overflow-hidden">
            <Sidebar />
            <div className="flex flex-col flex-grow overflow-hidden">
                <HeaderNav />
                <main className="flex-grow overflow-auto">
                    <Outlet />
                </main>
            </div>
        </div>
    );
}
