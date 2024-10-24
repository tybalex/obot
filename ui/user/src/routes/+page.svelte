<script lang="ts">
	import Navbar from '$lib/components/Navbar.svelte';
	import Drawer from '../lib/components/Drawer.svelte';
	import Messages from '$lib/components/Messages.svelte';
	import Editor from '$lib/components/Editor.svelte';
	import { loadFile } from '$lib/components/Editor.svelte';
	import SpeedDial from '$lib/components/SpeedDial.svelte';
	import Notifications from '$lib/components/Notifications.svelte';
	import { NotificationMessage } from '$lib/components/Notifications.svelte';
	import type { Input } from '$lib/components/messages/Input.svelte';
	import type { Messages as MessagesType } from '$lib/services';

	let editorVisible = false;
	let drawerVisible = false;
	let notification: Notifications;
	let messageDiv: HTMLDivElement;
	let messages: Messages;
	let editor: Editor;

	function handleError(event: CustomEvent<Error>) {
		notification.addNotification(new NotificationMessage(event.detail));
	}

	async function submit(e: CustomEvent<Input>) {
		return messages.submit(e.detail);
	}

	function handleLoadFile(e: CustomEvent<string>) {
		loadFile(e.detail);
		editorVisible = true;
		drawerVisible = false;
	}

	function handleMessages(e: CustomEvent<MessagesType>) {
		if (!messageDiv) {
			return;
		}

		// Check if messageDiv is already scrolled to the bottom
		let isScrolledToBottom =
			messageDiv.scrollHeight - messageDiv.clientHeight <= messageDiv.scrollTop + 1;

		const messages = e.detail;
		if (messages.messages.length > 0 && messages.messages[messages.messages.length - 1].sent) {
			// If the last message is a sent (user input) message, scroll to the bottom always
			isScrolledToBottom = true;
		}

		if (isScrolledToBottom) {
			setTimeout(() => {
				messageDiv.scrollTop = messageDiv.scrollHeight - messageDiv.clientHeight;
			}, 100);
		}
	}
</script>

<Navbar></Navbar>

<main id="main-content" class="flex h-screen justify-center">
	<Drawer bind:visible={drawerVisible} on:loadfile={handleLoadFile} />

	<div class="relative flex w-1/2 flex-1 justify-center">
		<div bind:this={messageDiv} class="w-full overflow-auto px-8 pb-32 pt-16 scrollbar-none">
			<div class="mx-auto max-w-[1000px]">
				<Messages
					bind:this={messages}
					on:focus={() => {
						drawerVisible = false;
					}}
					on:error={handleError}
					on:messages={handleMessages}
					on:loadfile={handleLoadFile}
				/>
			</div>
		</div>
	</div>

	{#if editorVisible}
		<div class="w-1/2 overflow-auto pb-16 pt-16 scrollbar-none">
			<Editor
				bind:this={editor}
				on:editor-close={() => {
					editorVisible = false;
				}}
				on:explain={submit}
				on:improve={submit}
			/>
		</div>
	{/if}

	<Notifications bind:this={notification} />
	<SpeedDial bind:drawerVisible bind:editorVisible />
</main>
