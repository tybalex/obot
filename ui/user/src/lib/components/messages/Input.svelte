<script module lang="ts">
	import type { Explain } from '$lib/services';

	export interface Input {
		prompt: string;
		explain?: Explain;
		improve?: Explain;
		changedFiles: Record<string, string>;
	}

	export interface Chat {
		submit: (override?: Input) => Promise<void>;
	}
</script>

<script lang="ts">
	import { ChatService } from '$lib/services';
	import { onMount } from 'svelte';
	import { editor } from '$lib/stores';
	import { autoHeight } from '$lib/actions/textarea';
	import { ArrowUp } from 'lucide-svelte';

	interface Props {
		onError?: (err: Error) => void;
		onFocus?: () => void;
		onSubmit?: (input: string | Input) => void | Promise<void>;
		readonly?: boolean;
		assistant: string;
	}

	let { onError, onFocus, onSubmit, assistant, readonly }: Props = $props();

	let value = $state('');
	let chat: HTMLTextAreaElement;

	function toInvokeInput(input: Input): Input | string {
		// This is just to make it pretty and send simple prompts if we can
		if (input.explain || input.improve) {
			return input;
		}
		if (input.changedFiles && Object.keys(input.changedFiles).length !== 0) {
			return input;
		}
		if (input.prompt) {
			return input.prompt;
		}
		return input;
	}

	export async function submit(override?: Input) {
		let input: Input = {
			prompt: value,
			changedFiles: {}
		};

		if (override) {
			input = override;
			if (chat) {
				chat.focus();
			}
		}

		for (const file of editor) {
			if (file.modified) {
				input.changedFiles[file.name] = file.buffer;
			}
		}

		try {
			if (readonly && !override) {
				await ChatService.abort(assistant);
			} else {
				const invokeInput = toInvokeInput(input);
				await onSubmit?.(invokeInput);
				await ChatService.invoke(assistant, invokeInput);
			}
		} catch (err) {
			if (err instanceof Error) {
				onError?.(err);
			} else {
				onError?.(new Error(String(err)));
			}
			return;
		}

		if (input.changedFiles) {
			for (const file of editor) {
				if (input.changedFiles[file.name]) {
					file.contents = input.changedFiles[file.name];
					file.modified = false;
					file.buffer = '';
				}
			}
		}

		value = '';
	}

	async function onKey(e: KeyboardEvent) {
		if (e.key !== 'Enter' || e.shiftKey) {
			return;
		}
		e.preventDefault();
		await submit();
	}

	function focusChat(e: KeyboardEvent) {
		if (!chat) {
			return;
		}

		const alphanumericRegex = /^[a-zA-Z0-9]$/;
		if (alphanumericRegex.test(e.key) && !e.ctrlKey && !e.altKey && !e.metaKey) {
			chat.focus();
		} else if ((e.key == 'Backspace' || e.key == 'ArrowLeft') && chat.value !== '') {
			chat.focus();
		}
	}

	onMount(() => {
		document.addEventListener('keydown', focusChat);
	});
</script>

<div
	class="absolute inset-x-0 bottom-0 z-30 flex justify-center bg-gradient-to-t from-white px-3 pb-8 pt-10 dark:from-black"
>
	<div class="w-full max-w-[700px]">
		<label for="chat" class="sr-only">Your message</label>
		<div
			class="flex items-center rounded-3xl
		bg-gray-70
		px-3
		focus-within:border-none
		focus-within:shadow-md
		focus-within:ring-1
		focus-within:ring-blue
		dark:border-none
		 dark:bg-gray-950"
		>
			<textarea
				use:autoHeight
				id="chat"
				rows="1"
				bind:value
				{readonly}
				onkeydown={onKey}
				bind:this={chat}
				onfocus={onFocus}
				class="peer
				ml-4
				mr-2
				 w-full resize-none
				 bg-gray-70 p-2.5 outline-none dark:bg-gray-950"
				placeholder="Your message..."
			></textarea>
			<button
				type="submit"
				onclick={() => submit()}
				class="rounded-full bg-gray-70 p-1
				text-blue
				hover:border-none
				hover:bg-gray-100
							 dark:bg-gray-950 dark:text-blue dark:hover:bg-gray-900"
			>
				{#if readonly}
					<div class="m-1.5 h-3 w-3 place-self-center rounded-sm bg-blue"></div>
				{:else}
					<ArrowUp />
				{/if}
				<span class="sr-only">Send message</span>
			</button>
		</div>
	</div>
</div>
