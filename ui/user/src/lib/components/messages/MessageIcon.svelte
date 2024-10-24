<script lang="ts">
	import Otto from '$lib/components/icons/Otto.svelte';
	import ProfileIcon from '$lib/components/profile/ProfileIcon.svelte';
	import type { Message } from '$lib/services';
	import Icon from '$lib/components/icons/Icon.svelte';
	import type { IconSource } from '@steeze-ui/svelte-icon';
	import { onMount } from 'svelte';

	export let msg: Message;
	let stockIcon: IconSource;

	onMount(async () => {
		if (msg.icon?.startsWith('stock:')) {
			const icons = await import(`@steeze-ui/heroicons`);
			stockIcon = icons[msg.icon.replace('stock:', '')] as IconSource;
		}
	});
</script>

{#if !msg.icon}
	<!-- Nothing -->
{:else if msg.icon.startsWith('stock:')}
	<Icon src={stockIcon} class="ml-1 mr-3 h-8 w-8" />
{:else if msg.icon === 'Otto'}
	<Otto />
{:else if msg.icon === 'Profile'}
	<ProfileIcon />
{:else}
	<img class="h-8 w-8 rounded-md bg-gray-100 p-1 text-blue-400" src={msg.icon} alt="message icon" />
{/if}
