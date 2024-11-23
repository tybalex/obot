<script lang="ts">
	import { Play } from 'lucide-svelte';
	import { ChatService, type TaskRun } from '$lib/services';
	import { onDestroy } from 'svelte';
	import Table from '$lib/components/tasks/Table.svelte';
	import { currentAssistant } from '$lib/stores';
	import { Trash } from '$lib/icons';

	interface Props {
		id: string;
	}

	let { id }: Props = $props();
	let runs: TaskRun[] = $state([]);
	let timeout: number;

	onDestroy(() => {
		if (timeout) {
			clearTimeout(timeout);
		}
	});

	currentAssistant.subscribe((assistant) => {
		if (assistant.id) {
			listRuns();
		}
	});

	async function listRuns() {
		try {
			if ($currentAssistant.id && id) {
				runs = (await ChatService.listTaskRuns($currentAssistant.id, id)).items;
			}
		} finally {
			timeout = setTimeout(listRuns, 15000);
		}
	}

	async function run() {
		if ($currentAssistant.id && id) {
			await ChatService.runTask($currentAssistant.id, id);
			runs = (await ChatService.listTaskRuns($currentAssistant.id, id)).items;
		}
	}

	async function deleteTask(runId: string) {
		if ($currentAssistant.id && id) {
			await ChatService.deleteTaskRun($currentAssistant.id, id, runId);
			runs = (await ChatService.listTaskRuns($currentAssistant.id, id)).items;
		}
	}
</script>

{#snippet runButton()}
	<button
		class="flex items-center gap-2 rounded-3xl bg-blue px-5 py-2 text-white hover:bg-blue-400"
		onclick={run}
	>
		Run now
		<Play class="h-5 w-5" />
	</button>
{/snippet}

{#snippet deleteButton(row: string[])}
	<button
		class="ms-3 text-gray-400 opacity-0 transition-opacity group-hover:opacity-100"
		onclick={() => deleteTask(row[2])}
	>
		<Trash class="h-5 w-5" />
	</button>
{/snippet}

{#if runs.length === 0}
	<div class="mt-8">
		{@render runButton()}
	</div>
{:else}
	<div class="mt-8">
		<div class="rounded-t-3xl bg-gray-50 px-5 pb-2 pt-5 dark:bg-gray-950">
			<h4 class="text-xl font-semibold">Runs</h4>
		</div>
		<Table
			header={['Date', 'Duration']}
			buttons={deleteButton}
			rows={runs.map((run) => {
				let endTime = run.endTime ? new Date(run.endTime) : new Date();
				let duration = 'Queued';
				if (run.startTime && run.endTime) {
					duration = `${Math.round((endTime.getTime() - new Date(run.startTime).getTime()) / 1000)}s`;
				} else if (run.startTime) {
					duration = `Running for ${Math.round((new Date().getTime() - new Date(run.startTime).getTime()) / 1000)}s`;
				}
				return [new Date(run.created).toTimeString(), duration, run.id];
			})}
		/>
		<div class="flex w-full justify-start rounded-b-3xl bg-gray-50 p-5 dark:bg-gray-950">
			{@render runButton()}
		</div>
	</div>
{/if}
