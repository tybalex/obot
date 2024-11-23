<script lang="ts">
	import { FileText, X } from '$lib/icons';
	import Milkdown from '$lib/components/editor/Milkdown.svelte';
	import Codemirror from '$lib/components/editor/Codemirror.svelte';
	import { EditorService } from '$lib/services';
	import Task from '$lib/components/tasks/Task.svelte';

	function fileChanged(e: CustomEvent<{ name: string; contents: string }>) {
		for (const item of EditorService.items) {
			if (item.name === e.detail.name) {
				item.buffer = e.detail.contents;
				item.modified = true;
			}
		}
	}
</script>

<div>
	<div class="flex items-center justify-between">
		<ul class="mb-4 flex flex-wrap text-center text-sm font-medium">
			{#each EditorService.items as item}
				<li class="me-2">
					<a
						href={`#editor:${item.name}`}
						class:selected={item.selected}
						onclick={() => {
							EditorService.select(item.id);
						}}
						class="selected active group flex items-center justify-center gap-2 rounded-t-lg p-4 text-black dark:border-blue dark:text-white"
						aria-current="page"
					>
						<FileText />
						<span>{item.name}</span>
						<button
							class="ml-2"
							onclick={() => {
								EditorService.remove(item.id);
							}}
						>
							<X />
						</button>
					</a>
				</li>
			{/each}
		</ul>
		<button
			class="icon-button"
			onclick={() => {
				EditorService.visible.set(false);
			}}
		>
			<X />
		</button>
	</div>

	<div
		id="editor"
		onkeydown={(e) => {
			e.stopPropagation();
		}}
		role="none"
		class="contents"
	>
		{#each EditorService.items as file}
			<div class:hidden={!file.selected} class="contents">
				{#if file.name.toLowerCase().endsWith('.md')}
					<Milkdown {file} on:changed={fileChanged} />
				{:else if file.task}
					<Task
						id={file.id}
						onChanged={(task) => {
							file.task = task;
							file.name = task.name || file.name;
						}}
					/>
				{:else}
					<Codemirror {file} on:changed={fileChanged} on:explain on:improve />
				{/if}
			</div>
		{/each}
	</div>
</div>

<style lang="postcss">
	.selected {
		@apply border-b-2 border-blue-600;
	}
</style>
