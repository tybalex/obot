<script lang="ts">
	import { Wrench } from '$lib/icons';
	import { tools } from '$lib/stores';
	import { currentAssistant } from '$lib/stores';
	import { ChatService } from '$lib/services';
	import Menu from '$lib/components/navbar/Menu.svelte';

	async function updateTool(enabled: boolean, tool: string | undefined) {
		if (!tool) {
			return;
		}
		if (enabled) {
			tools.set(await ChatService.enableTool($currentAssistant.id, tool));
		} else {
			tools.set(await ChatService.disableTool($currentAssistant.id, tool));
		}
	}
</script>

<Menu title="Tools" description="Enable or disable available tools">
	{#snippet icon()}
		<Wrench class="h-5 w-5" />
	{/snippet}
	{#snippet body()}
		<ul class="space-y-4 pt-6 text-sm">
			{#each $tools.items as tool, i}
				{#if !tool.builtin}
					<li>
						<div class="flex">
							{#if tool.icon}
								<img
									class="h-8 w-8 rounded-md bg-gray-100 p-1"
									src={tool.icon}
									alt="message icon"
								/>
							{:else}
								<Wrench class="h-8 w-8 rounded-md bg-gray-100 p-1 text-black" />
							{/if}
							<div class="flex flex-1 px-2">
								<div>
									<label for="checkbox-item-{i}" class="text-sm font-medium dark:text-gray-100"
										>{tool.name}</label
									>
									<p class="text-xs font-normal text-gray dark:text-gray-300">
										{tool.description}
									</p>
								</div>
							</div>
							<input
								id="checkbox-item-{i}"
								type="checkbox"
								checked={tool.enabled}
								onchange={(e) => updateTool(e.currentTarget.checked, tool.id)}
								disabled={tool.builtin || $tools.readonly}
								class="h-4 w-4 self-center"
							/>
						</div>
					</li>
				{/if}
			{/each}
		</ul>
	{/snippet}
</Menu>
