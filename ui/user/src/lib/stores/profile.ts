import { getProfile } from '$lib/services/chat/operations';
import { type Profile } from '$lib/services/chat/types';
import { storeWithInit } from './storeinit';
import { writable } from 'svelte/store';

const store = writable<Profile>({
	email: '',
	iconURL: '',
	role: 0
});

async function init() {
	try {
		store.set(await getProfile());
	} catch (e) {
		if (e instanceof Error && e.message.startsWith('403')) {
			store.set({
				email: '',
				iconURL: '',
				role: 0,
				unauthorized: true
			});
		} else {
			setTimeout(init, 5000);
		}
	}
}
export default storeWithInit(store, init);
