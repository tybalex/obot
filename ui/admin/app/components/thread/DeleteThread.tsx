import { TrashIcon } from "lucide-react";

import { ThreadsService } from "~/lib/service/api/threadsService";

import { ConfirmationDialog } from "~/components/composed/ConfirmationDialog";
import { Button } from "~/components/ui/button";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "~/components/ui/tooltip";
import { useConfirmationDialog } from "~/hooks/component-helpers/useConfirmationDialog";
import { useAsync } from "~/hooks/useAsync";

type DeleteThreadProps = { id: string };

export function DeleteThread({ id }: DeleteThreadProps) {
	const deleteThread = useAsync(ThreadsService.deleteThread, {
		onSuccess: () => ThreadsService.getThreads.revalidate(),
	});

	const { interceptAsync, dialogProps } = useConfirmationDialog();

	return (
		<>
			<Tooltip>
				<TooltipContent>Delete</TooltipContent>
				<TooltipTrigger
					asChild
					onClick={() => interceptAsync(() => deleteThread.executeAsync(id))}
				>
					<Button variant="ghost" size="icon">
						<TrashIcon />
					</Button>
				</TooltipTrigger>
			</Tooltip>

			<ConfirmationDialog
				{...dialogProps}
				title="Delete Thread"
				content="Are you sure you want to delete this thread? This action cannot be undone."
				confirmProps={{ variant: "destructive", children: "Delete" }}
			/>
		</>
	);
}
