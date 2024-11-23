import { newMessageEventSource } from './operations';
import type { Progress } from './types';

export function newMessageSource(
	assistant: string,
	onProgress: (message: Progress) => void,
	opts?: {
		task?: {
			id: string;
			follow?: boolean;
		};
		onError?: (error: Error) => void;
		onClose?: () => void;
	}
): () => void {
	let replayComplete = false;

	const es = newMessageEventSource(assistant, {
		task: opts?.task
	});
	es.onmessage = handleMessage;
	es.onopen = () => {
		console.log('Message EventSource opened');
	};
	es.addEventListener('close', () => {
		console.log('Message EventSource closed');
		opts?.onClose?.();
		es.close();
	});
	es.onerror = (e: Event) => {
		if (e.eventPhase === EventSource.CLOSED) {
			console.log('Message EventSource closed');
		}
	};

	function handleMessage(event: MessageEvent) {
		const progress = JSON.parse(event.data) as Progress;
		if (progress.replayComplete) {
			replayComplete = true;
		}
		if (progress.error) {
			if (progress.error.includes('abort')) {
				onProgress(progress);
			} else if (replayComplete && opts?.onError) {
				opts.onError(new Error(progress.error));
			}
		} else {
			onProgress(progress);
		}
	}

	return () => {
		if (es.readyState !== EventSource.CLOSED) {
			es.close();
		}
	};
}
