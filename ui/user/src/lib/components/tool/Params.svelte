<script lang="ts">
	import { Plus, Minus } from 'lucide-svelte';
	import { autoHeight } from '$lib/actions/textarea';

	interface Props {
		params: { key: string; value: string }[];
	}

	let { params = $bindable([]) }: Props = $props();
</script>

<div class="flex flex-col gap-4 rounded-3xl bg-gray-50 p-5 dark:bg-gray-950">
	<div class="flex">
		<h4 class="flex-1 text-xl font-semibold">Arguments</h4>
		<button onclick={() => params.push({ key: '', value: '' })}>
			<Plus class="h-5 w-5" />
		</button>
	</div>
	{#if params.length !== 0}
		<table class="w-full table-auto text-left">
			<thead>
				<tr>
					<th>Name</th>
					<th>Description</th>
				</tr>
			</thead>
			<tbody>
				{#each params as param, i}
					<tr>
						<td
							><input
								bind:value={param.key}
								placeholder="Enter name"
								class="ast bg-gray-50 outline-none dark:bg-gray-950"
							/></td
						>
						<td
							><textarea
								use:autoHeight
								class="resize-none bg-gray-50 outline-none dark:bg-gray-950"
								rows="1"
								bind:value={param.value}
							></textarea></td
						>
						<td>
							<button onclick={() => params.splice(i, 1)}>
								<Minus class="h-5 w-5" />
							</button>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>
