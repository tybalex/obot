<script lang="ts">
	import Trash from '$lib/components/icons/Trash.svelte';
	import { ChatService, type KnowledgeFile } from '$lib/services';
	import Loading from '$lib/components/icons/Loading.svelte';
	import { createEventDispatcher } from 'svelte';
	import { popover } from '$lib/actions';
	import { DocumentText, XCircle } from '@steeze-ui/heroicons';
	import Icon from '$lib/components/icons/Icon.svelte';

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
		<Icon class="h-4 text-red-500" src={XCircle} />
		<div class="rounded-md bg-red-500 p-2 text-white" use:tt.tooltip>
			{file.ingestionStatus?.error}
		</div>
	{:else if file.ingestionStatus?.status !== 'finished' && file.ingestionStatus?.status !== 'skipped'}
		<Loading />
	{/if}
	<Icon class="mr-2 text-gray-500 dark:text-gray-500" src={DocumentText} />
	<span class="flex-1 text-sm text-black dark:text-white">{file.fileName}</span>
	<button on:click={deleteFile}>
		<Trash class="hidden group-hover:block group-hover:text-red-700" />
	</button>
</div>
