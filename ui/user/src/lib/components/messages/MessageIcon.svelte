<script lang="ts">
	import Otto from '$lib/icons/Otto.svelte';
	import ProfileIcon from '$lib/components/profile/ProfileIcon.svelte';
	import type { Message } from '$lib/services';
	import { Pencil } from '$lib/icons';

	export let msg: Message;
</script>

{#if !msg.icon}
	<!-- Nothing -->
{:else if msg.icon === 'stock:Pencil'}
	<Pencil class="ml-1 mr-3 h-8 w-8" />
{:else if msg.icon === 'Otto'}
	<Otto />
{:else if msg.icon === 'Profile'}
	<ProfileIcon />
{:else}
	<img class="h-8 w-8 rounded-md bg-gray-100 p-1 text-blue-400" src={msg.icon} alt="message icon" />
{/if}
