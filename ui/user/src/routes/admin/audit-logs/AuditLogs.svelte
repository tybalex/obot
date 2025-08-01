<script lang="ts">
	import { twMerge } from 'tailwind-merge';
	import { tooltip } from '$lib/actions/tooltip.svelte';

	let {
		data = [],
		onSelectRow,
		emptyContent,
		fetchUserById,
		currentFragmentIndex = 0,
		getFragmentIndex,
		getFragmentRowIndex,
		onLoadNextFragment
	} = $props();
</script>

<!-- Data Table -->
<div
	class="dark:bg-surface2 w-full overflow-hidden overflow-x-auto rounded-lg border border-transparent bg-white shadow-sm"
>
	{#if data.length}
		<table class="min-w-full divide-y divide-gray-200">
			<thead class="dark:bg-surface1 bg-surface2">
				<tr class="sticky top-0">
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Timestamp</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>User</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Server</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Type</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Identifier</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Response Code</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Response Time (ms)</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>Client</th
					>
					<th
						scope="col"
						class="sticky top-0 px-6 py-3 text-left text-xs font-medium tracking-wider text-gray-500 uppercase"
						>IP Address</th
					>
				</tr>
			</thead>

			<tbody class="">
				<!-- Audit Data Rows -->
				{#each data as item, i (item.id)}
					{@const fragmentIndex = getFragmentIndex?.(i)}
					{@const fragmentRowIndex = getFragmentRowIndex?.(i)}
					<tr
						class={twMerge(
							'border-surface2 dark:border-surface2 border-t shadow-xs transition-colors duration-300',
							onSelectRow && ' hover:bg-surface1 dark:hover:bg-surface3 cursor-pointer',
							fragmentIndex && fragmentRowIndex === 0 && 'bg-surface3/50'
						)}
						data-fragment-index={fragmentIndex}
						data-fragment-row-index={fragmentRowIndex}
						onclick={() => onSelectRow?.(item)}
						{@attach (node) => {
							if (fragmentIndex < currentFragmentIndex) return;
							if (fragmentRowIndex > 0) return;

							const rootElement = document.body;

							const observer = new IntersectionObserver(
								(entries) => {
									const isIntersection = entries.some(
										(entry) => entry.target === node && entry.isIntersecting
									);

									if (isIntersection) {
										onLoadNextFragment?.(fragmentIndex);
									}
								},
								{
									root: rootElement
								}
							);

							observer.observe(node);

							return () => {
								observer.disconnect();
							};
						}}
					>
						<td class="px-6 py-4 text-sm whitespace-nowrap"
							>{new Date(item.createdAt)
								.toLocaleString(undefined, {
									year: 'numeric',
									month: 'short',
									day: 'numeric',
									hour: '2-digit',
									minute: '2-digit',
									second: '2-digit',
									hour12: true,
									timeZoneName: 'short'
								})
								.replace(/,/g, '')}</td
						>
						<td class="px-6 py-4 text-sm whitespace-nowrap">
							{#await fetchUserById(item.userID)}
								<span class="text-gray-500">Loading...</span>
							{:then user}
								{user?.displayName || 'Unknown User'}
							{/await}
						</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.mcpServerDisplayName}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.callType}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.callIdentifier}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.responseStatus}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.processingTimeMs}</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">
							<div class="max-w-[10ch] truncate" use:tooltip={item.client?.name}>
								{item.client?.name}
							</div>
						</td>
						<td class="px-6 py-4 text-sm whitespace-nowrap">{item.clientIP}</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{:else}
		{@render emptyContent?.()}
	{/if}
</div>
