<script lang="ts">
	import { fade } from 'svelte/transition';

	export let onClick: (() => void) | null = null;
	export let error: Error;
</script>

<!--error component-->
<div in:fade class="flex items-center justify-center">
	<div class="text-center">
		<h3 class="text-2xl font-semibold text-black dark:text-white">Error</h3>
		<p class="text-gray-500 dark:text-gray-400">{error.message}</p>
		{#if onClick}
			<button class="mt-4 text-black hover:underline dark:text-white" on:click={onClick}
				>Try again</button
			>
		{/if}
	</div>
</div>
