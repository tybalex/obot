<script lang="ts">
	import { Plus } from '$lib/icons';
	import { Pencil } from '$lib/icons';
	import { onMount } from 'svelte';

	export let drawerVisible = false;
	export let editorVisible = false;

	onMount(() => {
		if (window.location.href.indexOf('#editor') > -1) {
			editorVisible = true;
		}
	});
</script>

<div class="group fixed bottom-6 end-6 z-50">
	<div class="mb-4 hidden flex-col items-center space-y-2 group-hover:flex">
		<button
			type="button"
			on:click={() => {
				editorVisible = !editorVisible;
				if (!editorVisible) {
					window.location.href = '#';
				}
			}}
			class="flex h-[52px] w-[52px] items-center justify-center rounded-full border
						 border-gray-200 bg-white text-gray-500 shadow-sm hover:bg-gray-50 hover:text-gray-900
						  focus:outline-none focus:ring-4 focus:ring-gray-300 dark:border-gray-600 dark:bg-gray-700 dark:text-gray-400
						   dark:hover:bg-gray-600 dark:hover:text-white dark:focus:ring-gray-400"
		>
			<Pencil class="h-5 w-5" />
			<span class="sr-only">Editor</span>
		</button>
	</div>
	<button
		type="button"
		on:click={() => {
			drawerVisible = !drawerVisible;
		}}
		class="flex h-14 w-14 items-center justify-center rounded-full bg-blue-700 text-white hover:bg-blue-800 focus:outline-none focus:ring-4 focus:ring-blue-300 dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
	>
		<Plus class="transition-transform group-hover:rotate-45" />
		<span class="sr-only">Open actions menu</span>
	</button>
</div>
