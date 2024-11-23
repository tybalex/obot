<script lang="ts">
	import { Upload } from '$lib/icons';
	import { ChatService } from '$lib/services';
	import type { KnowledgeFile } from '$lib/services';
	import Loading from '$lib/icons/Loading.svelte';
	import Error from '$lib/components/Error.svelte';
	import { currentAssistant } from '$lib/stores';

	interface Props {
		onUpload?: () => void | Promise<void>;
	}

	let { onUpload }: Props = $props();

	let files = $state<FileList>();
	let uploadInProgress = $state<Promise<KnowledgeFile>>();

	$effect(() => {
		if (!files?.length) {
			return;
		}

		uploadInProgress = ChatService.uploadKnowledge($currentAssistant.id, files[0]);
		uploadInProgress
			.then(() => {
				onUpload?.();
			})
			.catch((error) => {
				console.error(error);
			})
			.finally(() => {
				uploadInProgress = undefined;
			});

		files = undefined;
	});
</script>

<div class="flex justify-end">
	<label
		class="-mb-3 -mr-3 mt-3 flex cursor-pointer justify-end gap-2 rounded-3xl p-3 px-4 hover:bg-gray-500 hover:text-white"
	>
		Upload (.pdf, .pptx, .txt)
		<input bind:files type="file" class="hidden" accept=".pdf, .txt, .doc, .docx" />
		{#await uploadInProgress}
			<Loading class="h-5 w-5" />
		{:catch error}
			<Error {error} />
		{/await}
		{#if !uploadInProgress}
			<Upload class="h-5 w-5" />
		{/if}
	</label>
</div>
