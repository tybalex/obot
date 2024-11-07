<script lang="ts">
	import { assistants } from '$lib/stores';

	$effect(() => {
		for (const assistant of $assistants) {
			if (assistant.current && assistant.id) {
				window.location.href = `/${assistant.id}`;
			}
		}
	});
</script>
