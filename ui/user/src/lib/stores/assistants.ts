import { page } from '$app/stores';
import { listAssistants } from '$lib/services/chat/operations';
import { type Assistant } from '$lib/services/chat/types';
import { storeWithInit } from '$lib/stores/storeinit';
import { writable } from 'svelte/store';

function assignSelected(assistants: Assistant[], selectedName: string): Assistant[] {
	const result: Assistant[] = [];

	for (const assistant of assistants) {
		assistant.current = selectedName !== '' && assistant.id === selectedName;
		result.push(assistant);
	}

	return result;
}

const store = writable<Assistant[]>(assignSelected([], ''));

export default storeWithInit(store, async () => {
	page.subscribe(async (value) => {
		const selectedName = value.params?.agent ?? '';
		try {
			const assistants = await listAssistants();
			store.set(assignSelected(assistants.items, selectedName));
		} catch {
			// just ignore
		}
	});
});
