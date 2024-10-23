import { TypographyP } from "~/components/Typography";
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from "~/components/ui/tooltip";

export function TruncatedText({
    content,
    maxWidth,
}: {
    content: string;
    maxWidth: string;
}) {
    return (
        <TooltipProvider>
            <Tooltip>
                <TooltipTrigger asChild>
                    <div className={`${maxWidth} truncate cursor-pointer`}>
                        <TypographyP className="truncate">
                            {content}
                        </TypographyP>
                    </div>
                </TooltipTrigger>
                <TooltipContent>
                    <p>{content}</p>
                </TooltipContent>
            </Tooltip>
        </TooltipProvider>
    );
}
