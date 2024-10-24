<script>
	import { MoonSolid, SunSolid } from 'flowbite-svelte-icons';
	import { darkMode } from '$lib/stores';
</script>

<button
	type="button"
	on:click={() => {
		darkMode.set(!$darkMode);
	}}
	class="icon-button"
>
	{#if $darkMode}
		<SunSolid class="h-5 w-5" />
	{:else}
		<MoonSolid class="h-5 w-5" />
	{/if}
</button>
