<script>
	import { ChatService } from '$lib/services';
	import Error from '$lib/components/Error.svelte';
	import Loading from '$lib/icons/Loading.svelte';
	import KnowledgeFile from '$lib/components/drawer/KnowledgeFile.svelte';
	import KnowledgeUpload from '$lib/components/drawer/KnowledgeUpload.svelte';
	import { onMount } from 'svelte';
	import { Brain } from '$lib/icons';

	let req = ChatService.getKnowledgeFiles();

	function reGet() {
		const promise = ChatService.getKnowledgeFiles();
		promise.finally(() => {
			req = promise;
		});
	}

	onMount(() => {
		const interval = setInterval(reGet, 5000);
		return () => clearInterval(interval);
	});
</script>

<div>
	<hr class="my-8 h-px border-0 bg-gray-200 dark:bg-gray-700" />
	<h5 class="mb-4 inline-flex items-center text-base font-semibold text-black dark:text-gray-100">
		<Brain class="mr-2 h-4 w-4" />
		Knowledge
	</h5>
	<p class="mb-6 text-sm text-black dark:text-gray-100">Searchable content</p>

	{#await req}
		<div class="flex items-center justify-center gap-1 dark:text-white">
			<Loading />
			Loading...
		</div>
	{:then files}
		<div>
			{#each files.items as file}
				<KnowledgeFile {file} on:deleted={reGet} />
			{/each}
		</div>
	{:catch error}
		<Error {error} onClick={reGet} />
	{/await}

	<KnowledgeUpload on:uploaded={reGet} />
</div>
