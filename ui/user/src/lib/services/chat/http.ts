// import from stores/errors and not stores to avoid a circular dependency
import errors from '$lib/stores/errors';

export let baseURL = 'http://localhost:8080/api';

if (typeof window !== 'undefined') {
	baseURL = baseURL.replace('http://localhost:8080', window.location.origin);
}

interface GetOptions {
	text?: boolean;
}

export async function doGet(path: string, opts?: GetOptions): Promise<unknown> {
	const resp = await fetch(baseURL + path);
	if (!resp.ok) {
		const body = await resp.text();
		const e = new Error(`${resp.status} ${path}: ${body}`);
		errors.append(e);
		throw e;
	}

	if (opts?.text) {
		return await resp.text();
	}

	return await resp.json();
}

export async function doDelete(path: string): Promise<void> {
	const resp = await fetch(baseURL + path, {
		method: 'DELETE'
	});
	if (!resp.ok) {
		const body = await resp.text();
		const e = new Error(`${resp.status} ${path}: ${body}`);
		errors.append(e);
		throw e;
	}
}

export async function doPost(path: string, input: string | object | Blob): Promise<unknown> {
	let contentType = 'text/plain';
	if (input instanceof Blob) {
		contentType = 'application/octet-stream';
	} else if (typeof input === 'object') {
		input = JSON.stringify(input);
		contentType = 'application/json';
	}
	const resp = await fetch(baseURL + path, {
		method: 'POST',
		headers: {
			'Content-Type': contentType
		},
		body: input
	});
	if (!resp.ok) {
		const body = await resp.text();
		const e = new Error(`${resp.status} ${path}: ${body}`);
		errors.append(e);
		throw e;
	}
	if (resp.headers.get('Content-Type') === 'application/json') {
		return resp.json();
	}
	return resp.text();
}
