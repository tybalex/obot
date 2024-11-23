import type { Assistant } from '$lib/services';
import assistants from './assistants';
import { storeWithInit } from './storeinit';
import { writable } from 'svelte/store';

const def: Assistant = {
	id: '',
	icons: {},
	current: false
};

const store = writable<Assistant>(def);

export default storeWithInit(store, async () => {
	assistants.subscribe(async (assistants) => {
		for (const assistant of assistants) {
			if (assistant.current) {
				store.set(assistant);
				return;
			}
		}
		store.set(def);
	});
});
