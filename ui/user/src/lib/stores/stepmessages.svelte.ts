import type { Messages, Progress } from '$lib/services';
import { buildMessagesFromProgress } from '$lib/services/chat/messages';
import { newMessageSource } from '$lib/services/chat/messagesource';
import { SvelteMap } from 'svelte/reactivity';

export interface StepMessages {
	messages: Map<string, Messages>;
	close: () => void;
}

export function createStepMessages(
	assistant: string,
	opts?: {
		task?: {
			id: string;
			follow?: boolean;
		};
		onClose?: () => void;
	}
): StepMessages {
	const progresses: Progress[] = [];
	const result: Map<string, Messages> = new SvelteMap();
	const close = newMessageSource(assistant, append, {
		task: opts?.task,
		onClose: () => {
			close?.();
			opts?.onClose?.();
		}
	});

	function append(p: Progress) {
		progresses.push(p);
		const newMessages = new Map<string, Progress[]>();
		let stepID: string | undefined;
		for (const progress of progresses) {
			if (progress.input) {
				continue;
			}
			if (progress.step?.id) {
				stepID = progress.step?.id.split('{')[0];
				newMessages.delete(stepID);
			}
			if (stepID) {
				if (!newMessages.has(stepID)) {
					newMessages.set(stepID, []);
				}
				newMessages.get(stepID)?.push(progress);
			}
		}

		for (const [stepID, msgs] of newMessages) {
			result.set(stepID, buildMessagesFromProgress(msgs));
		}
	}

	return {
		messages: result,
		close
	};
}
