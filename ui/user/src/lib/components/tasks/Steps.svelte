<script lang="ts">
	import {
		ChatService,
		type Messages,
		type Project,
		type Task,
		type TaskStep
	} from '$lib/services';
	import { onDestroy } from 'svelte';
	import Step from '$lib/components/tasks/Step.svelte';
	import { LoaderCircle, OctagonX, Play } from 'lucide-svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import { Thread } from '$lib/services/chat/thread.svelte';
	import Input from '$lib/components/tasks/Input.svelte';
	import Type from '$lib/components/tasks/Type.svelte';
	import Files from '$lib/components/tasks/Files.svelte';
	import { errors } from '$lib/stores';

	interface Props {
		task: Task;
		save?: () => Promise<void>;
		project: Project;
	}

	let { task = $bindable(), save, project }: Props = $props();

	let stepMessages = new SvelteMap<string, Messages>();
	let allMessages = $state<Messages>({ messages: [], inProgress: false });
	let input = $state('');
	let error = $state('');
	let thread: Thread | undefined = $state<Thread>();
	let pending = $derived(thread?.pending ?? false);
	let running = $derived(allMessages.inProgress);

	onDestroy(closeThread);
	$effect(resetThread);

	function resetThread() {
		if (!thread) {
			newThread();
		}
		error = '';
	}

	function closeThread() {
		if (!thread) {
			return;
		}

		thread.close();
		thread = undefined;
		stepMessages.clear();
		allMessages = { messages: [], inProgress: false };
	}

	function newThread(runID?: string) {
		closeThread();
		thread = new Thread(project, {
			onError: errors.items.push,
			task: task,
			runID: runID
		});
		stepMessages.clear();
		thread.onStepMessages = (stepID, messages) => {
			stepMessages.set(stepID, messages);
		};
		thread.onMessages = (messages) => {
			allMessages = messages;
		};
	}

	async function click() {
		error = '';

		const hasAtLeastOneInstruction = task.steps.some((step) => (step.step ?? '').trim().length > 0);
		if (!hasAtLeastOneInstruction) {
			error = 'At least one instruction is required to run the task.';
			return;
		}

		if (running || pending) {
			return await ChatService.abort(project.assistantID, project.id, {
				taskID: task.id,
				runID: 'editor'
			});
		}
		await run();
	}

	async function run(step?: TaskStep) {
		error = '';
		if (!thread || !task.id) {
			return;
		}

		await save?.();

		await thread.runTask(task.id, {
			stepID: step?.id || '*',
			input: input
		});
	}
</script>

<Input bind:input {task} />

<div class="rounded-2xl bg-gray-50 p-5 dark:bg-gray-950">
	<div class="flex items-center justify-between">
		<h4 class="text-lg font-semibold">Steps</h4>
		<Type bind:task />
		<div>
			<button
				class="bg-blue ml-2 flex items-center gap-2 rounded-3xl px-5 py-2 text-white hover:bg-blue-400"
				onclick={click}
			>
				{#if running}
					Stop
					<OctagonX class="h-4 w-4" />
				{:else if pending}
					Cancel
					<LoaderCircle class="h-4 w-4 animate-spin" />
				{:else}
					Test
					<Play class="h-4 w-4" />
				{/if}
			</button>
		</div>
	</div>

	<ol class="list-decimal pt-2 opacity-100">
		{#if task.steps.length > 0}
			{#key task.steps[0].id}
				<Step
					{run}
					bind:task
					bind:step={task.steps[0]}
					index={0}
					{stepMessages}
					{pending}
					{project}
				/>
			{/key}
		{/if}
	</ol>

	{#if error}
		<div class="mt-2 text-red-500">{error}</div>
	{/if}
</div>

<Files taskID={task.id} runID="editor" running={running || pending} {project} />
