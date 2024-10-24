<script lang="ts">
	import { createEventDispatcher, onDestroy, onMount } from 'svelte';
	import { ChatService } from '$lib/services';
	import type { Progress } from '$lib/services';

	let es: EventSource;
	let replayComplete = false;
	const dispatch = createEventDispatcher();

	function disconnect() {
		if (es) {
			es.close();
		}
	}

	function connect() {
		disconnect();
		es = ChatService.newMessageEventSource();
		es.onmessage = onMessage;
		es.onerror = (e: Event) => {
			if (e.eventPhase === EventSource.CLOSED) {
				disconnect();
				console.log('connection closed');
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
				dispatch('error', new Error(message.error));
			}
		} else {
			dispatch('message', message);
		}
	}

	onMount(connect);
	onDestroy(disconnect);
</script>
