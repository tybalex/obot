import { BrainCircuit, Handshake, Rocket } from "lucide-react";

import { useChat } from "~/components/chat/ChatContext";
import { Button } from "~/components/ui/button";

export function NoMessages() {
    const { processUserMessage } = useChat();

    const handleAddMessage = (content: string) => {
        processUserMessage(content, "user");
    };

    return (
        <div className="flex flex-col items-center justify-center space-y-4 text-center p-4 h-full">
            <h2 className="text-2xl font-semibold">Start the conversation!</h2>
            <p className="text-gray-500">
                Looking for a starting point? Try one of these options.
            </p>
            <div className="flex flex-wrap justify-center gap-2">
                <Button
                    variant="secondary"
                    onClick={() => handleAddMessage("Hello, how are you?")}
                >
                    <Handshake className="w-4 h-4 mr-2" />
                    Greeting
                </Button>
                <Button
                    variant="secondary"
                    onClick={() => handleAddMessage("What can you do?")}
                >
                    <Rocket className="w-4 h-4 mr-2" />
                    Capabilities
                </Button>
                <Button
                    variant="secondary"
                    onClick={() =>
                        handleAddMessage(
                            "Using your knowledge tools, tell me about your knowledge set."
                        )
                    }
                >
                    <BrainCircuit className="w-4 h-4 mr-2" />
                    Knowledge
                </Button>
            </div>
        </div>
    );
}
