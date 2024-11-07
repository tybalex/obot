<script lang="ts">
	import { Upload } from '$lib/icons';
	import { ChatService } from '$lib/services/index.js';
	import type { KnowledgeFile } from '$lib/services';
	import Loading from '$lib/icons/Loading.svelte';
	import Error from '$lib/components/Error.svelte';
	import { createEventDispatcher } from 'svelte';

	let upload: HTMLInputElement = $state();
	let uploadInProgress: Promise<KnowledgeFile> | null = $state();
	let dispatch = createEventDispatcher();

	async function uploadFile() {
		if (!upload.files || upload.files.length === 0) return;
		uploadInProgress = ChatService.uploadKnowledge(upload.files[0]);
		try {
			await uploadInProgress;
			dispatch('uploaded');
		} finally {
			uploadInProgress = null;
		}
	}
</script>

<div class="flex w-full items-center justify-center">
	{#await uploadInProgress}
		<div class="flex h-32 w-full items-center justify-center gap-2 text-black dark:text-gray-100">
			<Loading />
			Uploading...
		</div>
	{:catch error}
		<Error {error} />
	{/await}
	<label
		for="dropzone-file"
		ondragover={(e) => e.preventDefault()}
		ondrop={(e) => {
			e.preventDefault();
			if (e.dataTransfer?.files) {
				upload.files = e.dataTransfer.files;
				uploadFile();
			}
		}}
		class:hidden={uploadInProgress}
		class="mt-4 flex h-32 w-full cursor-pointer flex-col items-center justify-center rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 hover:bg-gray-100 dark:border-gray-600 dark:bg-gray-700 dark:hover:border-gray-500 dark:hover:bg-gray-600"
	>
		<div class="flex flex-col items-center justify-center pb-6 pt-5">
			<Upload />
			<p class="mb-2 text-sm text-gray-500 dark:text-gray-400">
				<span class="font-semibold">Click to upload</span> or drag and drop
			</p>
			<p class="text-xs text-gray-500 dark:text-gray-400">PDF, TXT, or DOCX</p>
		</div>
		<input
			bind:this={upload}
			onchange={uploadFile}
			id="dropzone-file"
			type="file"
			class="hidden"
			accept=".pdf, .txt, .docs"
		/>
	</label>
</div>
