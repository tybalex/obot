import { ReactNode, createContext, useContext } from "react";
import useSWR from "swr";

import { Role, User } from "~/lib/model/users";
import { UserService } from "~/lib/service/api/userService";

export const AuthDisabledUsername = "nobody";

interface AuthContextType {
    me: User;
    isLoading: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const { data: me, isLoading } = useSWR(
        UserService.getMe.key(),
        () => UserService.getMe(),
        { fallbackData: { role: Role.Default } as User }
    );

    return (
        <AuthContext.Provider value={{ me, isLoading }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error("useAuth must be used within a AuthProvider");
    }
    return context;
}
