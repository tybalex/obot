<script lang="ts">
	import ProfileIcon from '$lib/components/profile/ProfileIcon.svelte';
	import { profile } from '$lib/stores';
	import { popover } from '$lib/actions';

	const { ref, tooltip, toggle } = popover({
		placement: 'bottom-end'
	});
</script>

<!-- Profile -->
<div class="ml-1 flex items-center" use:ref>
	<button
		onclick={toggle}
		type="button"
		class="flex rounded-full bg-gray-800 text-sm focus:ring-4 focus:ring-gray-300 dark:focus:ring-gray-600"
	>
		<span class="sr-only">Open user menu</span>
		<ProfileIcon />
	</button>
	<!-- Dropdown menu -->
	<div
		use:tooltip
		class="mt-2 list-none divide-y divide-gray-100 rounded bg-white text-base shadow dark:divide-gray-600 dark:bg-gray-700"
	>
		<div class="px-4 py-3" role="none">
			<p class="truncate text-sm font-medium text-gray-900 dark:text-white" role="none">
				{$profile.email || 'Anonymous'}
			</p>
		</div>
		<ul class="py-1" role="none">
			{#if $profile.role === 1}
				<li>
					<a
						href="/admin/"
						rel="external"
						class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-white dark:hover:bg-gray-600 dark:hover:text-white"
						role="menuitem">Settings</a
					>
				</li>
			{/if}
			<li>
				<a
					href="/oauth2/sign_out"
					rel="external"
					class="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100 dark:text-white dark:hover:bg-gray-600 dark:hover:text-white"
					role="menuitem">Sign out</a
				>
			</li>
		</ul>
	</div>
</div>
