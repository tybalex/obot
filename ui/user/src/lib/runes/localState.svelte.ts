import { untrack } from 'svelte';

function defaultParser<T = string>(raw?: string | null) {
	if (raw === undefined || raw === null) return '' as T;
	return raw as T;
}

type LocalStateParams<T = string> = {
	value?: T;
	parse?: (raw?: string | null) => T;
};
export function localState<T = string>(
	key: string,
	{ value: defaultVallue, parse = defaultParser }: LocalStateParams<T> = {}
) {
	let value = $state<T>();
	let isReady = $state(false);

	$effect(() => {
		const storedValue = localStorage.getItem(key);
		untrack(() => {
			if (storedValue) {
				value = parse(storedValue) as T;
			} else {
				value = defaultVallue ?? ('' as T);
			}

			isReady = true;
		});
	});

	$effect(() => {
		if (!localStorage) return;

		localStorage.setItem(key, JSON.stringify(value));
	});

	return {
		get current() {
			return value;
		},
		set current(v) {
			value = v as T;
		},
		get isReady() {
			return isReady;
		}
	};
}
