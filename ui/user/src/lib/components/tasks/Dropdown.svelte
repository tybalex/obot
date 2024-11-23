<script lang="ts">
	import { ChevronDown } from '$lib/icons';
	import { popover } from '$lib/actions';
	import { fade } from 'svelte/transition';

	interface Props {
		values: Record<string, string>;
		selected?: string;
		disabled?: boolean;
		onSelected?: (value: string) => void | Promise<void>;
	}

	const { ref, tooltip, toggle } = popover({
		placement: 'bottom-start'
	});
	let { values, selected, disabled = false, onSelected }: Props = $props();

	async function select(value: string) {
		await onSelected?.(value);
		toggle();
	}
</script>

<button
	use:ref
	onclick={() => {
		if (!disabled) {
			toggle();
		}
	}}
	class="flex items-center gap-2 capitalize"
>
	{selected ? values[selected] : values[''] || ''}
	{#if !disabled}
		<div transition:fade={{ duration: 300 }}>
			<ChevronDown />
		</div>
	{/if}
</button>

<div use:tooltip class="z-30 rounded-lg bg-white shadow dark:bg-gray-700">
	<ul>
		{#each Object.keys(values) as key}
			{@const value = values[key]}
			<li class:selected={selected === key}>
				<button class="w-full text-start capitalize" onclick={() => select(key)}>
					{value}
				</button>
			</li>
		{/each}
	</ul>
</div>

<style lang="postcss">
	button {
		@apply px-4 py-1 hover:bg-gray-100;
	}

	.selected {
		@apply bg-gray-50;
	}

	li:first-of-type {
		@apply rounded-t-lg pt-2;
	}

	li:last-of-type {
		@apply rounded-b-lg pb-2;
	}
</style>
