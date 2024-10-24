<script lang="ts">
	import { Trash, FileText, CircleX } from '$lib/icons';
	import { ChatService, type KnowledgeFile } from '$lib/services';
	import Loading from '$lib/icons/Loading.svelte';
	import { createEventDispatcher } from 'svelte';
	import { popover } from '$lib/actions';

	export let file: KnowledgeFile;
	let dispatch = createEventDispatcher();
	const tt = popover({
		hover: true
	});

	async function deleteFile() {
		await ChatService.deleteKnowledgeFile(file.fileName);
		dispatch('deleted');
	}
</script>

<div class="group flex cursor-pointer items-center rounded-md p-2" use:tt.ref>
	{#if file.ingestionStatus?.status === 'failed'}
		<CircleX class="h-4 text-red-500" />
		<div class="rounded-md bg-red-500 p-2 text-white" use:tt.tooltip>
			{file.ingestionStatus?.error}
		</div>
	{:else if file.ingestionStatus?.status !== 'finished' && file.ingestionStatus?.status !== 'skipped'}
		<Loading />
	{/if}
	<FileText class="mr-2 text-gray-500 dark:text-gray-500" />
	<span class="flex-1 text-sm text-black dark:text-white">{file.fileName}</span>
	<button on:click={deleteFile}>
		<Trash class="hidden group-hover:block group-hover:text-red-700" />
	</button>
</div>
