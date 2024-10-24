<script lang="ts">
	import { Icon } from '@steeze-ui/svelte-icon';
	import type { IconSource } from '@steeze-ui/svelte-icon';
	import { twMerge } from 'tailwind-merge';

	export let src: IconSource;
</script>

<Icon {src} class={twMerge('h-6 w-6 text-black dark:text-white', $$props.class)} />
