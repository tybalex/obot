<script lang="ts">
	import Self from './Step.svelte';
	import type { TaskStep } from '$lib/services';
	import type { StepMessages } from '$lib/stores';
	import Message from '$lib/components/messages/Message.svelte';
	import { X } from '$lib/icons';
	import { CheckCircle, RefreshCcw, Save, Undo, XCircle } from 'lucide-svelte';
	import { tick } from 'svelte';
	import { autoHeight } from '$lib/actions/textarea';

	interface Props {
		parentStale?: boolean;
		onChange?: (steps: TaskStep[]) => void | Promise<void>;
		run?: (step: TaskStep, steps?: TaskStep[]) => Promise<void>;
		steps: TaskStep[];
		index: number;
		editMode?: boolean;
		stepMessages?: StepMessages;
	}

	let {
		parentStale,
		onChange,
		run,
		steps,
		index,
		editMode = false,
		stepMessages
	}: Props = $props();

	let step = $derived(steps[index]);
	let messages = $derived(stepMessages?.messages.get(step.id)?.messages ?? []);
	let currentValue = $state(textValue(steps[index]));
	let dirty = $derived(textValue(steps[index]) !== currentValue);
	let stale: boolean = $derived(dirty || parentStale || false);
	let isIf: boolean = $derived(currentValue?.startsWith('If ') || false);
	let isThenPath: boolean = $derived.by(() => {
		for (const message of messages) {
			return message.message.join('').trim().toLowerCase() == 'true';
		}
		return false;
	});
	let isElsePath: boolean = $derived.by(() => {
		for (const message of messages) {
			return message.message.join('').trim().toLowerCase() == 'false';
		}
		return false;
	});

	$effect(() => {
		if (!currentValue?.trimStart().toLowerCase().startsWith('if ')) {
			return;
		}

		const newValue = currentValue.trimStart().replace(/^[iI][fF]/, 'If');
		if (newValue !== currentValue) {
			currentValue = newValue;
		}

		if (steps.length < index + 2) {
			console.log('if detected');
			steps.push(createStep());
		}

		if (!step.if) {
			console.log('creating if');
			step.if = {
				condition: '',
				steps: [createStep()],
				else: [createStep()]
			};
		}
	});

	function textValue(step: TaskStep) {
		return step.if ? 'If ' + step.if.condition : step.step;
	}

	async function deleteStep() {
		const newSteps = [...steps];
		newSteps.splice(index, 1);
		await onChange?.(newSteps);
	}

	async function revert() {
		if (dirty) {
			currentValue = step.step;
		}
	}

	function firstLine(e: KeyboardEvent) {
		return (
			e.target instanceof HTMLTextAreaElement &&
			e.target.value.lastIndexOf('\n', e.target.selectionStart - 1) === -1 &&
			e.target.selectionStart === e.target.selectionEnd
		);
	}

	function lastLine(e: KeyboardEvent) {
		return (
			e.target instanceof HTMLTextAreaElement &&
			e.target.value.indexOf('\n', e.target.selectionStart) === -1 &&
			e.target.selectionStart === e.target.selectionEnd
		);
	}

	function lastChar(e: KeyboardEvent) {
		return (
			e.target instanceof HTMLTextAreaElement &&
			e.target.selectionStart === e.target.value.length &&
			e.target.selectionStart === e.target.selectionEnd
		);
	}

	function synchronized(newSteps?: TaskStep[]): TaskStep[] | undefined {
		if (!newSteps && !dirty) {
			return;
		}

		const retSteps = newSteps ?? [...steps];
		if (dirty) {
			const newStep = { ...step };
			if (currentValue?.startsWith('If ')) {
				newStep.step = '';
				if (!newStep.if) {
					newStep.if = {
						condition: ''
					};
				}
				newStep.if.condition = currentValue.replace(/^If /, '').trim();
			} else {
				newStep.step = currentValue;
				newStep.if = undefined;
			}
			retSteps[index] = newStep;
		}

		return retSteps;
	}

	async function save(steps?: TaskStep[]) {
		const newSteps = synchronized(steps);
		if (newSteps) {
			await onChange?.(newSteps);
		}
	}

	async function onkeydown(e: KeyboardEvent) {
		if (e.key === 'Enter' && e.ctrlKey) {
			e.preventDefault();
			await doRun();
		} else if (e.key === 'ArrowUp' && firstLine(e) && e.target instanceof HTMLTextAreaElement) {
			const oldIndex = e.target.selectionStart;
			setTimeout(() => {
				if (e.target instanceof HTMLTextAreaElement && e.target.selectionStart === 0) {
					const prevNode = document.getElementById('step' + steps[index - 1]?.id);
					if (prevNode) {
						e.target.selectionStart = oldIndex;
						e.target.selectionEnd = oldIndex;
						prevNode.focus();
					}
				}
			});
		} else if (e.key === 'ArrowDown' && lastLine(e) && e.target instanceof HTMLTextAreaElement) {
			const oldIndex = e.target.selectionStart;
			setTimeout(() => {
				if (
					e.target instanceof HTMLTextAreaElement &&
					e.target.selectionStart === e.target.value.length
				) {
					const nextNode = document.getElementById('step' + steps[index + 1]?.id);
					if (nextNode) {
						e.target.selectionStart = oldIndex;
						e.target.selectionEnd = oldIndex;
						nextNode.focus();
					}
				}
			});
		} else if (
			e.key === 'Enter' &&
			!e.ctrlKey &&
			!e.shiftKey &&
			lastChar(e) &&
			e.target instanceof HTMLTextAreaElement &&
			e.target.value.trim() !== ''
		) {
			const newStep = createStep();
			const newSteps = [...steps];
			newSteps.splice(index + 1, 0, newStep);
			e.preventDefault();
			await save(newSteps);
			await tick();
			document.getElementById('step' + newStep.id)?.focus();
		} else if (
			e.key === 'Backspace' &&
			e.target instanceof HTMLTextAreaElement &&
			e.target.value === '' &&
			steps.length > 1
		) {
			e.preventDefault();
			await deleteStep();
			document.getElementById('step' + steps[index - 1]?.id)?.focus();
		}
	}

	function setIfSteps(ifSteps?: TaskStep[], elseSteps?: TaskStep[]): TaskStep[] {
		const newStep: TaskStep = {
			...step,
			if: {
				...(step.if ?? { condition: '' }),
				steps: ifSteps,
				else: elseSteps
			}
		};
		const newSteps = [...steps];
		newSteps[index] = newStep;
		return newSteps;
	}

	function createStep(): TaskStep {
		return { id: Math.random().toString(36).substring(7), step: '' };
	}

	async function doRun() {
		await run?.(step, synchronized());
	}
</script>

<li>
	<div class="flex items-center justify-between">
		{#if editMode}
			{#if isIf}
				<span class="keyword z-10 -mr-3 uppercase">If</span>
			{/if}
			<textarea
				{onkeydown}
				rows="1"
				placeholder="Stuff..."
				use:autoHeight
				id={'step' + step.id}
				bind:value={currentValue}
				class="flex-1 resize-none border-none bg-gray-50 outline-none dark:bg-gray-950"
			></textarea>
			<div class="flex gap-2 p-2">
				<button class="rounded-lg p-2 hover:bg-gray-50 dark:hover:bg-gray-950" onclick={doRun}>
					<RefreshCcw class="h-4 w-4" />
				</button>
				{#if dirty}
					<button
						class="-ml-3 rounded-lg p-2 hover:bg-gray-50 dark:hover:bg-gray-950"
						onclick={revert}
					>
						<Undo class="h-4 w-4" />
					</button>
					<button
						class="-ml-3 rounded-lg p-2 hover:bg-gray-50 dark:hover:bg-gray-950"
						onclick={async () => {
							await save();
						}}
					>
						<Save class="h-4 w-4" />
					</button>
				{/if}
				<button
					class="-ml-3 rounded-lg p-2 hover:bg-gray-50 dark:hover:bg-gray-950"
					onclick={deleteStep}
				>
					<X class="h-4 w-4" />
				</button>
			</div>
		{:else if isIf}
			<div>
				<span class="keyword uppercase">If</span>
				<span>{currentValue?.slice(3)}</span>
			</div>
		{:else}
			<span>{currentValue}</span>
		{/if}
	</div>
	{#if messages.length > 0}
		<div class="-mx-11 bg-white p-5">
			{#each messages as msg}
				<Message {msg} />
			{/each}
		</div>
	{/if}

	{#if steps.length > index + 1 && isIf}
		{#if step?.if?.steps?.length}
			{#key step.id + ' then'}
				<ol class="then-else">
					<span class="keyword">Then</span>
					{#if isThenPath}
						<CheckCircle class="h-5 w-5" />
					{/if}
					<Self
						onChange={async (steps) => {
							await onChange?.(setIfSteps(steps, step?.if?.else));
						}}
						run={async (step, steps) => {
							if (steps) {
								await run?.(step, setIfSteps(steps, step?.if?.else));
							} else {
								await run?.(step, steps);
							}
						}}
						{editMode}
						steps={step.if.steps}
						index={0}
						{stepMessages}
						parentStale={stale}
					/>
				</ol>
			{/key}
		{/if}
		{#if step?.if?.else?.length}
			{#key step.id + ' else'}
				<ol class="then-else">
					<span class="keyword">Else</span>
					{#if isElsePath}
						<XCircle class="h-5 w-5" />
					{/if}
					<Self
						onChange={async (steps) => {
							await onChange?.(setIfSteps(step?.if?.steps, steps));
						}}
						run={async (step, steps) => {
							if (steps) {
								await run?.(step, setIfSteps(step?.if?.steps, steps));
							} else {
								await run?.(step, steps);
							}
						}}
						{editMode}
						steps={step.if.else}
						index={0}
						{stepMessages}
						parentStale={stale}
					/>
				</ol>
			{/key}
		{/if}
	{/if}
</li>

{#if steps.length > index + 1}
	{#key steps[index + 1].id}
		<Self
			{onChange}
			{run}
			{editMode}
			{steps}
			index={index + 1}
			{stepMessages}
			parentStale={stale}
		/>
	{/key}
{/if}

<style lang="postcss">
	ol {
		@apply list-[lower-alpha] pl-2;
	}

	li {
		@apply ms-6;
	}

	li::marker {
		@apply font-semibold;
	}

	.keyword {
		@apply rounded-md bg-blue px-2 text-white shadow-md;
	}

	.then-else {
		@apply mb-2 mt-3 p-4;
	}
</style>
