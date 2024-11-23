<script lang="ts">
	import { ChatService, type Task, type TaskStep } from '$lib/services';
	import { createStepMessages, currentAssistant, type StepMessages } from '$lib/stores';
	import { onDestroy } from 'svelte';
	import Step from '$lib/components/tasks/Step.svelte';
	import { Play } from 'lucide-svelte';
	import { autoHeight } from '$lib/actions/textarea';

	interface Props {
		task: Task;
		editMode?: boolean;
		save?: (steps: TaskStep[]) => void | Promise<void>;
	}

	let { task, editMode = false, save }: Props = $props();

	let stepMessages = $state<StepMessages>();
	let steps = $derived(task.steps ?? []);
	let input = $state('');

	onDestroy(() => {
		if (stepMessages) {
			stepMessages.close();
		}
	});

	$effect(() => {
		if (!editMode && stepMessages) {
			stepMessages.close();
			stepMessages = undefined;
		}
	});

	function listen() {
		stepMessages = createStepMessages($currentAssistant.id, {
			task: {
				id: task.id,
				follow: true
			},
			onClose: () => {
				stepMessages = undefined;
				setTimeout(listen, 2000);
			}
		});
	}

	async function run(step?: TaskStep, steps?: TaskStep[]) {
		if (!stepMessages) {
			listen();
		}

		if (steps) {
			await save?.(steps);
		}

		const resp = await ChatService.runTask($currentAssistant.id, task.id, {
			stepID: step?.id || '*',
			input: input
		});

		console.log(`running task ${task.id} step ${step?.id} on runID ${resp.id}`);
	}
</script>

{#if editMode}
	<div class="mt-8 rounded-3xl bg-gray-50 p-5 dark:bg-gray-950">
		<h4 class="text-xl font-semibold">Test Input</h4>
		<textarea
			bind:value={input}
			use:autoHeight
			rows="1"
			class="mt-2 w-full resize-none rounded-md bg-gray-50 p-2 outline-none dark:bg-gray-950"
			placeholder="Enter input"
		></textarea>
	</div>
{/if}

<div class="mt-8 rounded-3xl bg-gray-50 p-5 dark:bg-gray-950">
	<div class="flex items-center justify-between">
		<h4 class="text-xl font-semibold">Steps</h4>
		{#if editMode}
			<button
				class="flex items-center gap-2 rounded-3xl bg-blue px-5 py-2 text-white hover:bg-blue-400"
				onclick={async () => {
					await run();
				}}
			>
				Run now
				<Play class="h-5 w-5" />
			</button>
		{/if}
	</div>

	<ol class="list-decimal pt-2 opacity-100">
		{#if steps.length > 0}
			{#key steps[0].id}
				<Step {run} onChange={save} {steps} index={0} {stepMessages} {editMode} />
			{/key}
		{/if}
	</ol>
</div>
