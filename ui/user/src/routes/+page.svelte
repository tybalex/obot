<script lang="ts">
	import { assistants } from '$lib/stores';
	import { goto } from '$app/navigation';

	$effect(() => {
		let id = $assistants.find((assistant) => assistant.id === 'otto')?.id;
		if (!id) {
			id = $assistants.find((assistant) => assistant.id !== '')?.id;
		}
		if (id) {
			goto(`/${id}`);
		}
	});
</script>
