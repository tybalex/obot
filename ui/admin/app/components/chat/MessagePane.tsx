import { ToolCall } from "~/lib/model/chatEvents";
import { Message as MessageType } from "~/lib/model/messages";
import { cn } from "~/lib/utils";

import { useChat } from "~/components/chat/ChatContext";
import { Message } from "~/components/chat/Message";
import { NoMessages } from "~/components/chat/NoMessages";
import { LoadingSpinner } from "~/components/ui/LoadingSpinner";
import { ScrollArea } from "~/components/ui/scroll-area";

interface MessagePaneProps {
    messages: MessageType[];
    className?: string;
    classNames?: {
        root?: string;
        messageList?: string;
    };
    generatingMessage?: Nullish<MessageType>;
    generatingTools?: ToolCall[];
}

export function MessagePane({
    messages,
    className,
    classNames = {},
    generatingMessage,
}: MessagePaneProps) {
    const { readOnly, isLoading } = useChat();

    const isEmpty = messages.length === 0 && !generatingMessage && !readOnly;

    return (
        <div className={cn("flex flex-col h-full", className, classNames.root)}>
            <ScrollArea
                startScrollAt="bottom"
                enableScrollTo="bottom"
                enableScrollStick="bottom"
                className={cn("h-full w-full relative", classNames.messageList)}
            >
                {isLoading && isEmpty ? (
                    <LoadingSpinner fillContainer />
                ) : isEmpty ? (
                    <NoMessages />
                ) : (
                    <div className="p-4 space-y-6 w-full">
                        {messages.map((message, i) => (
                            <Message key={i} message={message} />
                        ))}
                        {generatingMessage && (
                            <Message message={generatingMessage} />
                        )}
                    </div>
                )}
            </ScrollArea>
        </div>
    );
}
