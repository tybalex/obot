import { ChevronUpIcon } from "lucide-react";
import React, { useState } from "react";
import useSWR from "swr";

import { Thread } from "~/lib/model/threads";
import { ThreadsService } from "~/lib/service/api/threadsService";

import { LoadingSpinner } from "~/components/ui/LoadingSpinner";
import { Button } from "~/components/ui/button";
import {
    Command,
    CommandEmpty,
    CommandGroup,
    CommandInput,
    CommandItem,
    CommandList,
} from "~/components/ui/command";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "~/components/ui/popover";

interface PastThreadsProps {
    agentId: string;
    onThreadSelect: (threadId: string) => void;
}

export const PastThreads: React.FC<PastThreadsProps> = ({
    agentId,
    onThreadSelect,
}) => {
    const [open, setOpen] = useState(false);
    const {
        data: threads,
        error,
        isLoading,
        mutate,
    } = useSWR(ThreadsService.getThreadsByAgent.key(agentId), () =>
        ThreadsService.getThreadsByAgent(agentId)
    );

    const handleOpenChange = (newOpen: boolean) => {
        setOpen(newOpen);
        if (newOpen) {
            mutate();
        }
    };

    const handleThreadSelect = (threadId: string) => {
        onThreadSelect(threadId);
        setOpen(false);
    };

    return (
        <Popover open={open} onOpenChange={handleOpenChange}>
            <PopoverTrigger asChild>
                <Button variant="secondary" size="icon">
                    <ChevronUpIcon className="w-4 h-4" />
                </Button>
            </PopoverTrigger>
            <PopoverContent className="w-80 p-0">
                <Command className="flex-col-reverse">
                    <CommandInput placeholder="Search threads..." />
                    <CommandList>
                        <CommandEmpty>No threads found.</CommandEmpty>
                        {isLoading ? (
                            <div className="flex justify-center items-center h-20">
                                <LoadingSpinner size={24} />
                            </div>
                        ) : error ? (
                            <div className="text-center text-red-500 p-2">
                                Failed to load threads
                            </div>
                        ) : threads && threads.length > 0 ? (
                            <CommandGroup>
                                {threads.map((thread: Thread) => (
                                    <CommandItem
                                        key={thread.id}
                                        onSelect={() =>
                                            handleThreadSelect(thread.id)
                                        }
                                        className="cursor-pointer"
                                    >
                                        <div>
                                            <h3 className="font-semibold">
                                                Thread
                                                <span className="ml-2 text-muted-foreground">
                                                    {thread.id}
                                                </span>
                                            </h3>
                                            <p className="text-sm text-gray-500">
                                                {new Date(
                                                    thread.created
                                                ).toLocaleString()}
                                            </p>
                                        </div>
                                    </CommandItem>
                                ))}
                            </CommandGroup>
                        ) : null}
                    </CommandList>
                </Command>
            </PopoverContent>
        </Popover>
    );
};
