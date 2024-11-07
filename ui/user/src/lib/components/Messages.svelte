<script lang="ts">
	import Input from '$lib/components/messages/Input.svelte';
	import type { Input as InputType } from '$lib/components/messages/Input.svelte';
	import MessageSource from '$lib/components/messages/MessageSource.svelte';
	import type { Messages, Progress } from '$lib/services';
	import { ChatService } from '$lib/services';
	import Message from '$lib/components/messages/Message.svelte';

	interface Props {
		assistant: string;
		onmessages: (messages: Messages) => void;
		onerror: (err: Error) => void;
		onfocus?: () => void;
		onloadfile: (filename: string) => void;
	}

	let {
		assistant,
		onmessages,
		onerror,
		onfocus,
		onloadfile
	}: Props = $props()

	let progressEvents: Progress[] = [];
	let replayComplete = false;
	let messages: Messages = $state({ messages: [], inProgress: false });

	function handleMessage(progress: Progress) {
		progressEvents = [...progressEvents, progress];
		if (!replayComplete) {
			replayComplete = progressEvents.find((e) => e.replayComplete) !== undefined;
		}

		if (!replayComplete) {
			return;
		}

		messages = ChatService.progressToMessages(progressEvents);

		// forward the messages to the parent component
		onmessages(messages)
	}

	let inputBox: ReturnType<typeof Input>;

	export async function submit(input: InputType) {
		return inputBox.submit(input);
	}
</script>

<MessageSource {assistant} onmessage={handleMessage} {onerror} />

<div class="flex flex-col gap-8">
	{#each messages.messages as msg}
		{#if !msg.ignore}
			<Message {msg} {onloadfile} />
		{/if}
	{/each}
	<Input {assistant} bind:this={inputBox} {onerror} {onfocus} />
</div>
