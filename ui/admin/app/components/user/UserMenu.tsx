import { User } from "lucide-react";
import React from "react";

import { roleToString } from "~/lib/model/users";
import { cn } from "~/lib/utils";

import { Avatar, AvatarFallback, AvatarImage } from "~/components/ui/avatar";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "~/components/ui/popover";

import { useAuth } from "../auth/AuthContext";
import { Button } from "../ui/button";

interface UserMenuProps {
    className?: string;
}

export const UserMenu: React.FC<UserMenuProps> = ({ className }) => {
    const { me } = useAuth();

    if (me.username === "nobody") {
        return null;
    }

    return (
        <Popover>
            <PopoverTrigger asChild>
                <div
                    className={cn(
                        "flex items-center cursor-pointer",
                        className
                    )}
                >
                    <Avatar className="mr-4">
                        <AvatarImage />
                        <AvatarFallback>
                            <User className="w-5 h-5" />
                        </AvatarFallback>
                    </Avatar>
                    <div className="truncate max-w-full">
                        <p className="text-sm font-medium truncate">
                            {me?.email}
                        </p>
                        <p className="text-muted-foreground text-left text-xs truncate">
                            {roleToString(me?.role)}
                        </p>
                    </div>
                </div>
            </PopoverTrigger>
            <PopoverContent className="w-auto" side="bottom">
                <Button
                    variant="destructive"
                    onClick={() => {
                        window.location.href = "/oauth2/sign_out?rd=/admin/";
                    }}
                >
                    Sign Out
                </Button>
            </PopoverContent>
        </Popover>
    );
};
