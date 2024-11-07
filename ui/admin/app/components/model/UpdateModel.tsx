import { PenSquareIcon } from "lucide-react";
import { useState } from "react";

import { Model } from "~/lib/model/models";

import { ModelForm } from "~/components/model/ModelForm";
import { Button } from "~/components/ui/button";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogTitle,
    DialogTrigger,
} from "~/components/ui/dialog";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "~/components/ui/tooltip";

type UpdateModelProps = {
    model: Model;
};

export function UpdateModel(props: UpdateModelProps) {
    const { model } = props;
    const [open, setOpen] = useState(false);

    return (
        <TooltipProvider>
            <Tooltip>
                <Dialog open={open} onOpenChange={setOpen}>
                    <DialogContent>
                        <DialogTitle>Update model</DialogTitle>

                        <DialogDescription hidden>
                            Update model
                        </DialogDescription>

                        <ModelForm
                            model={model}
                            onSubmit={() => setOpen(false)}
                        />
                    </DialogContent>

                    <DialogTrigger asChild>
                        <TooltipTrigger asChild>
                            <Button size={"icon"} variant="ghost">
                                <PenSquareIcon />
                            </Button>
                        </TooltipTrigger>
                    </DialogTrigger>
                </Dialog>

                <TooltipContent>Update Model</TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}
