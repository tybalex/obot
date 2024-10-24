import { type Files, type KnowledgeFile, type KnowledgeFiles, type Profile } from '$lib/services';
import { baseURL, doDelete, doGet, doPost } from '$lib/services/chat/http';

export async function getProfile(): Promise<Profile> {
	const obj = (await doGet('/me')) as Profile;
	obj.isAdmin = () => {
		return obj.role === 1;
	};
	return obj;
}

export async function deleteKnowledgeFile(filename: string) {
	return doDelete(`/threads/user/knowledge/${filename}`);
}

export async function deleteFile(filename: string) {
	return doDelete(`/threads/user/files/${filename}`);
}

export async function getFile(filename: string): Promise<string> {
	return (await doGet(`/threads/user/file/${filename}`, {
		text: true
	})) as string;
}

export async function uploadKnowledge(file: File): Promise<KnowledgeFile> {
	return (await doPost(`/threads/user/knowledge/${file.name}`, file)) as KnowledgeFile;
}

interface DeletedItems<T extends Deleted> {
	items: T[];
}

interface Deleted {
	deleted?: string;
}

function removedDeleted<V extends Deleted, T extends DeletedItems<V>>(items: T): T {
	items.items = items.items?.filter((item) => !item.deleted);
	return items;
}

export async function getKnowledgeFiles(): Promise<KnowledgeFiles> {
	const files = (await doGet('/threads/user/knowledge')) as KnowledgeFiles;
	if (!files.items) {
		files.items = [];
	}
	return removedDeleted(files);
}

export async function getFiles(): Promise<Files> {
	const files = (await doGet('/threads/user/files')) as Files;
	if (!files.items) {
		files.items = [];
	}
	return files;
}

export async function invoke(msg: string | object) {
	await doPost('/invoke/otto/threads/user?async=true', msg);
}

export function newMessageEventSource(): EventSource {
	return new EventSource(baseURL + '/threads/user/events?waitForThread=true&follow=true');
}
