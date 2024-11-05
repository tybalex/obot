import {
    ReactNode,
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useState,
} from "react";
import { mutate } from "swr";

import { ChatEvent } from "~/lib/model/chatEvents";
import { Message, promptMessage, toolCallMessage } from "~/lib/model/messages";
import { InvokeService } from "~/lib/service/api/invokeService";
import { ThreadsService } from "~/lib/service/api/threadsService";

import { useAsync } from "~/hooks/useAsync";

type Mode = "agent" | "workflow";

interface ChatContextType {
    messages: Message[];
    mode: Mode;
    processUserMessage: (text: string, sender: "user" | "agent") => void;
    id: string;
    threadId: string | undefined;
    invoke: (prompt?: string) => void;
    readOnly?: boolean;
    isRunning: boolean;
    isInvoking: boolean;
}

const ChatContext = createContext<ChatContextType | undefined>(undefined);

export function ChatProvider({
    children,
    id,
    mode = "agent",
    threadId,
    onCreateThreadId,
    readOnly,
}: {
    children: ReactNode;
    mode?: Mode;
    id: string;
    threadId?: string;
    onCreateThreadId?: (threadId: string) => void;
    readOnly?: boolean;
}) {
    /**
     * processUserMessage is responsible for adding the user's message to the chat and
     * triggering the agent to respond to it.
     */
    const processUserMessage = (text: string, sender: "user" | "agent") => {
        if (mode === "workflow" || readOnly) return;
        const newMessage: Message = { text, sender };

        // insertMessage(newMessage);
        handlePrompt(newMessage.text);
    };

    const invoke = (prompt?: string) => {
        if (prompt && mode === "agent" && !readOnly) {
            handlePrompt(prompt);
        }
    };

    const handlePrompt = (prompt: string) => {
        if (prompt && mode === "agent" && !readOnly) {
            invokeAgent.execute({
                slug: id,
                prompt: prompt,
                thread: threadId,
            });
        }
        // do nothing if the mode is workflow
    };

    const invokeAgent = useAsync(InvokeService.invokeAgentWithStream, {
        onSuccess: ({ threadId: responseThreadId }) => {
            if (responseThreadId && !threadId) {
                // persist the threadId
                onCreateThreadId?.(responseThreadId);

                // revalidate threads
                mutate(ThreadsService.getThreads.key());
            }
        },
    });

    const { messages, isRunning } = useMessageSource(threadId);

    return (
        <ChatContext.Provider
            value={{
                messages,
                processUserMessage,
                mode,
                id,
                threadId,
                invoke,
                isRunning,
                isInvoking: invokeAgent.isLoading,
                readOnly,
            }}
        >
            {children}
        </ChatContext.Provider>
    );
}

export function useChat() {
    const context = useContext(ChatContext);
    if (context === undefined) {
        throw new Error("useChat must be used within a ChatProvider");
    }
    return context;
}

function useMessageSource(threadId?: string) {
    const [messageMap, setMessageMap] = useState<Map<string, Message>>(
        new Map()
    );
    const [isRunning, setIsRunning] = useState(false);

    const addContent = useCallback((event: ChatEvent) => {
        console.log(event);

        const { content, prompt, toolCall, runComplete, input, error, runID } =
            event;

        setIsRunning(!runComplete);

        setMessageMap((prev) => {
            const copy = new Map(prev);

            const contentID = event.contentID ?? crypto.randomUUID();

            const existing = copy.get(contentID);
            if (existing) {
                copy.set(contentID, {
                    ...existing,
                    text: existing.text + content,
                });

                return copy;
            }

            if (error) {
                copy.set(contentID, {
                    sender: "agent",
                    text: error,
                    runId: runID,
                    error: true,
                });
                return copy;
            }

            if (input) {
                copy.set(contentID, {
                    sender: "user",
                    text: input,
                    runId: runID,
                });
                return copy;
            }

            if (toolCall) {
                copy.set(contentID, toolCallMessage(toolCall));
                return copy;
            }

            if (prompt) {
                copy.set(contentID, promptMessage(prompt, runID));
                return copy;
            }

            if (content) {
                copy.set(contentID, {
                    sender: "agent",
                    text: content,
                    runId: runID,
                });
                return copy;
            }

            return copy;
        });
    }, []);

    useEffect(() => {
        setMessageMap(new Map());

        if (!threadId) return;

        const source = ThreadsService.getThreadEventSource(threadId);

        source.onmessage = (event) => {
            const chunk = JSON.parse(event.data) as ChatEvent;
            addContent(chunk);
        };

        return () => {
            source.close();
        };
    }, [threadId, addContent]);

    const messages = useMemo(() => {
        return Array.from(messageMap.values());
    }, [messageMap]);

    return { messages, messageMap, isRunning };
}
