<script lang="ts">
	import File from '$lib/components/drawer/File.svelte';
	import { ChatService } from '$lib/services';
	import Error from '$lib/components/Error.svelte';
	import Loading from '$lib/components/icons/Loading.svelte';

	let req = ChatService.getFiles();

	function reGet() {
		req = ChatService.getFiles();
	}
</script>

<div>
	<hr class="my-8 h-px border-0 bg-gray-200 dark:bg-gray-700" />
	<h5 class="mb-4 inline-flex items-center text-base font-semibold text-black dark:text-white">
		<svg
			class="me-2.5 h-4 w-4"
			aria-hidden="true"
			xmlns="http://www.w3.org/2000/svg"
			fill="none"
			viewBox="0 0 24 24"
		>
			<path
				stroke="currentColor"
				stroke-linejoin="round"
				stroke-width="2"
				d="M10 3v4a1 1 0 0 1-1 1H5m14-4v16a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V7.914a1 1 0 0 1 .293-.707l3.914-3.914A1 1 0 0 1 9.914 3H18a1 1 0 0 1 1 1Z"
			/>
		</svg>
		Files
	</h5>
	<p class="mb-6 text-sm text-black dark:text-gray-100">Editable files</p>

	{#await req}
		<div class="flex items-center justify-center gap-1 dark:text-white">
			<Loading />
			Loading...
		</div>
	{:then files}
		<div class="flex flex-col">
			{#each files.items as file}
				<File {file} on:deleted={reGet} on:loadfile />
			{/each}
		</div>
	{:catch error}
		<Error {error} onClick={reGet} />
	{/await}
</div>
