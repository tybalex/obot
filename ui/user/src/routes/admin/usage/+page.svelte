<script lang="ts">
	import { fade } from 'svelte/transition';
	import Layout from '$lib/components/Layout.svelte';
	import { PAGE_TRANSITION_DURATION } from '$lib/constants';
	import UsageGraphs from '$lib/components/admin/usage/UsageGraphs.svelte';

	const duration = PAGE_TRANSITION_DURATION;
</script>

<Layout>
	<div
		class="my-4 flex h-dvh min-h-full flex-col gap-8 pb-8"
		in:fade={{ duration }}
		out:fade={{ duration }}
	>
		<h1 class="text-2xl font-semibold">Usage</h1>
		<UsageGraphs />
	</div>
</Layout>

<svelte:head>
	<title>Obot | Usage</title>
</svelte:head>
