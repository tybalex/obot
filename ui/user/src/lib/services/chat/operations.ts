import {
	type Assistants,
	type Files,
	type KnowledgeFile,
	type KnowledgeFiles,
	type Profile
} from './types';
import { baseURL, doDelete, doGet, doPost } from './http';

export async function getProfile(): Promise<Profile> {
	const obj = (await doGet('/me')) as Profile;
	obj.isAdmin = () => {
		return obj.role === 1;
	};
	return obj;
}

export async function listAssistants(): Promise<Assistants> {
	const assistants = (await doGet(`/assistants`)) as Assistants;
	if (!assistants.items) {
		assistants.items = [];
	}
	return assistants;
}

export async function deleteKnowledgeFile(assistant: string, filename: string) {
	return doDelete(`/assistants/${assistant}/knowledge/${filename}`);
}

export async function deleteFile(assistant: string, filename: string) {
	return doDelete(`/assistants/${assistant}/files/${filename}`);
}

export async function getFile(assistant: string, filename: string): Promise<string> {
	return (await doGet(`/assistants/${assistant}/file/${filename}`, {
		text: true
	})) as string;
}

export async function uploadKnowledge(assistant: string, file: File): Promise<KnowledgeFile> {
	return (await doPost(`/assistants/${assistant}/knowledge/${file.name}`, file)) as KnowledgeFile;
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

export async function getKnowledgeFiles(assistant: string): Promise<KnowledgeFiles> {
	const files = (await doGet(`/assistants/${assistant}/knowledge`)) as KnowledgeFiles;
	if (!files.items) {
		files.items = [];
	}
	return removedDeleted(files);
}

export async function listFiles(assistant: string): Promise<Files> {
	const files = (await doGet(`/assistants/${assistant}/files`)) as Files;
	if (!files.items) {
		files.items = [];
	}
	return files;
}

export async function invoke(assistant: string, msg: string | object) {
	await doPost(`/assistants/${assistant}/invoke`, msg);
}

export function newMessageEventSource(assistant: string): EventSource {
	return new EventSource(baseURL + `/assistants/${assistant}/events`);
}
