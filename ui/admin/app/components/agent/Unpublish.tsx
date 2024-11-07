import { EyeOff } from "lucide-react";

import { Agent } from "~/lib/model/agents";

import { useAgent } from "~/components/agent/AgentContext";
import { ConfirmationDialog } from "~/components/composed/ConfirmationDialog";
import { Button } from "~/components/ui/button";

type PublishProps = {
    className?: string;
    onChange: (agent: Agent) => void;
};

export function Unpublish({ onChange }: PublishProps) {
    const { agent } = useAgent();

    return (
        <ConfirmationDialog
            title="Unpublish Agent"
            description="Are you sure you want to unpublish this agent? This action will disrupt every user currently using this reference."
            onConfirm={() => {
                onChange({
                    ...agent,
                    refName: "",
                });
            }}
            confirmProps={{
                variant: "destructive",
                children: "Unpublish",
            }}
        >
            <Button variant="secondary" size="sm">
                <EyeOff className="w-4 h-4" />
                Unpublish
            </Button>
        </ConfirmationDialog>
    );
}
