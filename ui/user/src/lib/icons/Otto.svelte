<script lang="ts">
	import { twMerge } from 'tailwind-merge';
	interface Props {
		class?: string;
	}
	const { class: klass = '' }: Props = $props();
</script>

<img src="/images/otto8-icon-blue.svg" class={twMerge('h-8 w-8', klass)} alt="otto8 Icon" />
