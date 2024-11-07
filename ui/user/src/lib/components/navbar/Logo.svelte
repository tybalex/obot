<script lang="ts">
	import { Check, ChevronDown } from '$lib/icons';
	import { assistants } from '$lib/stores';
	import { darkMode } from '$lib/stores';
	import type { Assistant } from '$lib/services';
	import { popover } from '$lib/actions';
	import { fade } from 'svelte/transition';

	const selected = $derived($assistants.find(a => a.current));

	const { ref, tooltip, toggle } = popover({
		placement: 'bottom-start'
	});

	function icon(a: Assistant | undefined): string {
		if (!a) {
			return '';
		}

		if ($darkMode) {
			return (a.icons.iconDark ? a.icons.iconDark : a.icons.icon) ?? '';
		}
		return a.icons.icon ?? '';
	}

	function collapsedIcon(a: Assistant | undefined): string {
		if (!a) {
			return '';
		}

		if ($darkMode) {
			return a.icons.collapsedDark ?? a.icons.iconDark ?? a.icons.collapsed ?? a.icons.icon ?? '';
		}
		return a.icons.collapsed ?? a.icons.icon ?? '';
	}

</script>

<div class="flex items-center justify-start" transition:fade|global >
	<a use:ref href={`/${selected?.id ?? ''}`}
		 class="text-purple-950 flex items-center gap-2" onclick={() => {
		if ($assistants.length > 1) {
			toggle();
		} else {
			window.location.href = `/${selected?.id ?? ''}`;
		}
	}}>
		{#if collapsedIcon(selected)}
			<img src={collapsedIcon(selected)} alt="assistant icon" class="ml-3 h-8" />
		{:else if selected?.name}
			<div class="w-8 h-8 rounded-full bg-gray-200 dark:bg-gray-500 flex items-center justify-center">{
				selected?.name ? selected?.name[0].toUpperCase() : '?'
			}
			</div>
			<span class="font-semibold dark:text-gray-100" >{selected?.name ?? ''}</span>
		{/if}
	</a>

	<!-- Dropdown menu -->
	<div use:tooltip
			 class="mt-4 bg-white divide-y divide-gray-100 rounded-lg shadow w-60 dark:bg-gray-700 dark:divide-gray-600">
		<ul class="p-3 space-y-1 text-sm text-gray-700 dark:text-gray-200" aria-labelledby="dropdownHelperButton">
			{#each $assistants as assistant}
				<li>
					<a href={'/' + assistant.id} data-sveltekit-reload
						 class="flex p-2 rounded hover:bg-gray-100 dark:hover:bg-gray-600">
						<div class="flex items-center h-5">
							{#if icon(assistant)}
								<img src={icon(assistant)} alt="assistant icon" class="w-5 h-5 rounded-full" />
							{:else}
								<div class="w-5 h-5 rounded-full bg-gray-200 dark:bg-gray-500 flex items-center justify-center">{
									assistant.name ? assistant.name[0].toUpperCase() : '?'
								}</div>
							{/if}
						</div>
						<div class="ms-2 text-sm">
							<label for="helper-checkbox-1" class="font-medium text-gray-900 dark:text-white flex items-center gap-1">
								{assistant.name}
							</label>
							<p class="text-xs font-normal text-gray-500 dark:text-gray-300">{assistant.description}</p>
						</div>
						<div class="flex-1 flex justify-end">
							{#if assistant.current}
								<Check class="w-4 h-4 self-center" />
							{/if}
						</div>
					</a>
				</li>
			{/each}
		</ul>
	</div>
	{#if $assistants.length > 1}
		<button class="h-full" onclick={toggle}>
			<ChevronDown
				class="w-5 h-5 ms-2 text-gray-200 dark:text-gray-700 hover:text-black hover:bg-gray-100 hover:dark:bg-gray-700 hover:dark:text-white rounded" />
		</button>
	{/if}
</div>
