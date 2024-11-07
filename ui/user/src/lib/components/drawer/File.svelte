<!-- @migration-task Error while migrating Svelte code: `<button>` is invalid inside `<button>` -->
<script lang="ts">
	import { Trash, FileText } from '$lib/icons';
	import { ChatService, type File as FileType } from '$lib/services';
	import { createEventDispatcher } from 'svelte';

	export let file: FileType;
	let dispatch = createEventDispatcher();

	async function deleteFile(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		await ChatService.deleteFile(file.name);
		dispatch('deleted');
	}
</script>

<a
	class="group flex cursor-pointer items-center rounded-md p-2 hover:bg-gray-400 hover:text-black"
	on:click={() => {
		dispatch('loadfile', file.name);
	}}
>
	<FileText class="mr-2 text-gray-500 dark:text-gray-500" />
	<span class="inline-flex flex-1 text-sm text-black group-hover:text-white dark:text-gray-100">
		{file.name}
	</span>
	<button on:click={deleteFile}>
		<Trash class="hidden group-hover:block group-hover:text-red-700" />
	</button>
</a>
