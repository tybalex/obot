<script lang="ts">
	import Files from '$lib/components/drawer/Files.svelte';
	import Knowledge from '$lib/components/drawer/Knowledge.svelte';
	import { InfoCircleSolid } from 'flowbite-svelte-icons';
	import Close from '$lib/components/icons/Close.svelte';

	export let visible = true;
</script>

<!-- drawer component -->
{#if visible}
	<button class="fixed inset-0 z-30 cursor-default" on:click={() => (visible = false)} />
	<div
		class="fixed top-0 z-40 h-screen w-1/3 translate-x-full overflow-y-auto bg-gray-50
	 p-4 transition-transform dark:bg-gray-800"
		tabindex="-1"
	>
		<h5 class="mb-4 inline-flex items-center text-base font-semibold text-black dark:text-white">
			<InfoCircleSolid class="me-2.5 h-5 w-5" />
			Content
		</h5>

		<button
			type="button"
			on:click={() => (visible = false)}
			class="absolute end-2.5 top-2.5 inline-flex h-8
				w-8 items-center justify-center rounded-lg bg-transparent text-sm text-gray-400 hover:bg-gray-200 hover:text-gray-900
				 dark:hover:bg-gray-600 dark:hover:text-white"
		>
			<Close />
			<span class="sr-only">Close menu</span>
		</button>

		<p class="mb-6 text-sm text-black dark:text-white">Manage your content</p>
		<Files on:loadfile />
		<Knowledge />
	</div>
{/if}
