<script lang="ts">
	import Input from '$lib/components/messages/Input.svelte';
	import type { Input as InputType } from '$lib/components/messages/Input.svelte';
	import MessageSource from '$lib/components/messages/MessageSource.svelte';
	import type { Messages, Progress } from '$lib/services';
	import { ChatService } from '$lib/services';
	import Message from '$lib/components/messages/Message.svelte';
	import { createEventDispatcher } from 'svelte';

	let progressEvents: Progress[] = [];
	let replayComplete = false;
	let messages: Messages = { messages: [], inProgress: false };
	const dispatcher = createEventDispatcher();

	function handleMessage(event: CustomEvent<Progress>) {
		progressEvents = [...progressEvents, event.detail];
		if (!replayComplete) {
			replayComplete = progressEvents.find((e) => e.replayComplete) !== undefined;
		}

		if (!replayComplete) {
			return;
		}

		messages = ChatService.progressToMessages(progressEvents);
		// forward the messages to the parent component
		dispatcher('messages', messages);
	}

	let inputBox: Input;

	export async function submit(input: InputType) {
		return inputBox.submit(input);
	}
</script>

<MessageSource on:message={handleMessage} on:error />

<div class="flex flex-col gap-8">
	{#each messages.messages as msg}
		{#if !msg.ignore}
			<Message {msg} on:loadfile />
		{/if}
	{/each}
	<Input bind:this={inputBox} on:error on:focus />
</div>
