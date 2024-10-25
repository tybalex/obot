import { Outlet, isRouteErrorResponse, useRouteError } from "@remix-run/react";
import { AxiosError } from "axios";

import { AuthDisabledUsername, useAuth } from "~/components/auth/AuthContext";
import { Error, RouteError, Unauthorized } from "~/components/errors";
import { HeaderNav } from "~/components/header/HeaderNav";
import { Sidebar } from "~/components/sidebar";
import { SignIn } from "~/components/signin/SignIn";

export function ErrorBoundary() {
    const error = useRouteError();
    const { me } = useAuth();

    switch (true) {
        case error instanceof AxiosError:
            if (
                error.response?.status === 403 &&
                me.username &&
                me.username !== AuthDisabledUsername
            ) {
                return <Unauthorized />;
            }
            return <SignIn />;
        case isRouteErrorResponse(error):
            return <RouteError error={error} />;
        default:
            return <Error error={error as Error} />;
    }
}

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
