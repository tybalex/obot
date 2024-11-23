import { listTools } from '$lib/services/chat/operations';
import { type AssistantToolList } from '$lib/services/chat/types';
import assistants from '$lib/stores/assistants';
import { storeWithInit } from '$lib/stores/storeinit';
import { writable } from 'svelte/store';

const store = writable<AssistantToolList>({
	readonly: true,
	items: []
});

export default storeWithInit(store, async () => {
	assistants.subscribe(async (assistants) => {
		for (const assistant of assistants) {
			if (assistant.current && assistant.id) {
				store.set(await listTools(assistant.id));
				break;
			}
		}
	});
});
