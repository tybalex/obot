<script lang="ts">
	import Trash from '$lib/components/icons/Trash.svelte';
	import { ChatService, type File as FileType } from '$lib/services';
	import { createEventDispatcher } from 'svelte';
	import { DocumentText } from '@steeze-ui/heroicons';
	import Icon from '$lib/components/icons/Icon.svelte';

	export let file: FileType;
	let dispatch = createEventDispatcher();

	async function deleteFile(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		await ChatService.deleteFile(file.name);
		dispatch('deleted');
	}
</script>

<button
	class="group flex cursor-pointer items-center rounded-md p-2 hover:bg-gray-400 hover:text-black"
	on:click={() => {
		dispatch('loadfile', file.name);
	}}
>
	<Icon class="mr-2 text-gray-500 dark:text-gray-500" src={DocumentText} />
	<span class="inline-flex flex-1 text-sm text-black group-hover:text-white dark:text-gray-100">
		{file.name}
	</span>
	<button on:click={deleteFile}>
		<Trash class="hidden group-hover:block group-hover:text-red-700" />
	</button>
</button>
