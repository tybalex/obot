import { type Assistant, ChatService } from '$lib/services';
import { writable } from 'svelte/store';
import { storeWithInit } from '$lib/stores/storeinit';
import { page } from '$app/stores';

function assignSelected(assistants: Assistant[], selectedName: string): Assistant[] {
	let found = false;
	const result: Assistant[] = [];

	for (const assistant of assistants) {
		if (selectedName !== '' && assistant.id === selectedName) {
			assistant.current = true;
			found = true;
		} else {
			assistant.current = false;
		}
		result.push(assistant);
	}

	if (!found && result.length > 0) {
		result[0].current = true;
	} else if (!found) {
		result.push({
			id: '',
			icons: {},
			current: false
		});
	}

	return result;
}

const store = writable<Assistant[]>(assignSelected([], ''));

export default storeWithInit(store, async () => {
	page.subscribe(async (value) => {
		const selectedName = value.params?.agent ?? '';
		if (selectedName) {
			const assistants = await ChatService.listAssistants();
			store.set(assignSelected(assistants.items, selectedName));
		}
	});
});
