import { type Assistant, ChatService } from '$lib/services';
import { writable } from 'svelte/store';
import { storeWithInit } from '$lib/stores/storeinit';
import { page } from '$app/stores';

let selectedName = '';

function assignSelected(assistants: Assistant[]): Assistant[] {
	let found = false;
	const result: Assistant[] = [];

	for (const assistant of assistants) {
		if (assistant.id === selectedName) {
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
			current: true
		});
	}

	return result;
}

const store = writable<Assistant[]>(assignSelected([]));

export default storeWithInit(store, async () => {
	page.subscribe((value) => {
		selectedName = value.params?.agent ?? '';
		store.update((assistants) => {
			return assignSelected(assistants);
		});
	});

	const assistants = await ChatService.listAssistants();
	store.set(assignSelected(assistants.items));
});
