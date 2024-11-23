<script lang="ts">
	import { ChatService, type Task, type Version } from '$lib/services';
	import Dropdown from '$lib/components/tasks/Dropdown.svelte';
	import Schedule from '$lib/components/tasks/Schedule.svelte';
	import { onMount } from 'svelte';
	import { Copy } from 'lucide-svelte';
	import OnDemand from '$lib/components/tasks/OnDemand.svelte';

	interface Props {
		task?: Task;
		onChanged?: (task: Task) => void | Promise<void>;
		editMode?: boolean;
	}

	let {
		task = {
			name: 'Loading...',
			steps: [],
			id: ''
		},
		onChanged,
		editMode = false
	}: Props = $props();

	let version: Version = $state({});
	let email = $derived.by(() => {
		if (version.emailDomain && task.alias) {
			return `${task.alias}@${version.emailDomain}`;
		}
		return '';
	});
	let webhook = $derived.by(() => {
		if (typeof window !== 'undefined' && task.alias) {
			return window.location.protocol + '//' + window.location.host + '/api/webhook/' + task.alias;
		}
		return '';
	});
	let options = $derived.by(() => {
		const options: Record<string, string> = {
			schedule: 'on interval',
			webhook: 'on webhook',
			onDemand: 'on demand'
		};
		if (version.emailDomain) {
			options['email'] = 'on email';
		}
		// assigned later so it's rendered last
		options['onDemand'] = 'on demand';
		return options;
	});
	let lastParamsSeen: Record<string, string> = $state({});

	$effect(() => {
		if (task.onDemand?.params && Object.keys(task.onDemand.params).length > 0) {
			lastParamsSeen = task.onDemand.params;
		}
		if (
			Object.keys(lastParamsSeen ?? {}).length > 0 &&
			task.onDemand?.params &&
			Object.keys(task.onDemand.params).length === 0
		) {
			lastParamsSeen = task.onDemand.params;
		}
	});

	onMount(async () => {
		version = await ChatService.getVersion();
	});

	function selectedTrigger(): string {
		if (task.schedule) {
			return 'schedule';
		}
		if (task.webhook) {
			return 'webhook';
		}
		if (task.email) {
			return 'email';
		}
		return 'onDemand';
	}

	async function selected(value: string) {
		if (value === 'schedule') {
			await onChanged?.({
				...task,
				schedule: {
					interval: 'hourly',
					hour: 0,
					minute: 0,
					day: 0,
					weekday: 0
				},
				webhook: undefined,
				email: undefined,
				onDemand: undefined
			});
		}
		if (value === 'webhook') {
			await onChanged?.({
				...task,
				schedule: undefined,
				webhook: {},
				email: undefined,
				onDemand: undefined
			});
		}
		if (value === 'email') {
			await onChanged?.({
				...task,
				schedule: undefined,
				webhook: undefined,
				onDemand: undefined,
				email: {}
			});
		}
		if (value === 'onDemand') {
			await onChanged?.({
				...task,
				schedule: undefined,
				webhook: undefined,
				email: undefined,
				onDemand: undefined
			});
		}
	}
</script>

{#if editMode || selectedTrigger() !== 'onDemand' || Object.keys(task?.onDemand?.params ?? {}).length > 0}
	<div class="mt-8 rounded-3xl bg-gray-50 p-5 dark:bg-gray-950">
		<div class="flex items-center justify-between">
			{#if editMode}
				<h4 class="text-xl font-semibold">Trigger</h4>
				<div class="flex">
					<Dropdown
						disabled={!editMode}
						selected={selectedTrigger()}
						values={options}
						onSelected={selected}
					/>
				</div>
			{:else}
				<h4 class="text-xl font-semibold capitalize">{options[selectedTrigger()]}</h4>
			{/if}
			{#if selectedTrigger() === 'schedule'}
				<Schedule
					schedule={task.schedule}
					{editMode}
					onChanged={async (schedule) => {
						await onChanged?.({
							...task,
							schedule
						});
					}}
				/>
			{/if}
		</div>
		{#if selectedTrigger() === 'webhook'}
			<div class="mt-3 flex justify-between pr-5">
				URL
				<div class="flex">
					{webhook}
					<Copy class="ml-2 h-5 w-5" />
				</div>
			</div>
		{/if}
		{#if selectedTrigger() === 'email' && email}
			<div class="mt-3 flex justify-between pr-5">
				{#if editMode}Email{/if}
				Address
				<div class="flex">
					{email}
					<Copy class="ml-2 h-5 w-5" />
				</div>
			</div>
		{/if}
		{#if selectedTrigger() === 'onDemand'}
			<OnDemand
				{editMode}
				onDemand={task.onDemand ?? {}}
				onChanged={async (onDemand) => {
					await onChanged?.({
						...task,
						onDemand
					});
				}}
			/>
		{/if}
	</div>
{/if}
