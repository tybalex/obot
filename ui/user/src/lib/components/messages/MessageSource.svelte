<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { ChatService } from '$lib/services';
	import type { Progress } from '$lib/services';

	interface Props {
		assistant: string;
		onmessage?: (event: Progress) => void;
		onerror?: (event: Error) => void;
	}

	let {
		assistant,
		onmessage = () => {},
		onerror = () => {},
	} : Props = $props();

	let es: EventSource;
	let replayComplete = false;

	function disconnect() {
		if (es) {
			es.close();
		}
	}

	function connect() {
		disconnect();
		es = ChatService.newMessageEventSource(assistant);
		es.onmessage = onMessage;
		es.onerror = (e: Event) => {
			if (e.eventPhase === EventSource.CLOSED) {
				disconnect();
				setTimeout(connect, 5000);
				return;
			}
		};
	}

	function onMessage(event: MessageEvent) {
		const message = JSON.parse(event.data) as Progress;
		if (message.replayComplete) {
			replayComplete = true;
		}
		if (message.error) {
			if (replayComplete) {
				onerror(new Error(message.error));
			}
		} else {
			onmessage(message);
		}
	}

	onMount(connect);
	onDestroy(disconnect);
</script>
