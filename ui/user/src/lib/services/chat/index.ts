import { buildMessagesFromProgress } from '$lib/services/chat/messages';
import {
	getProfile,
	listFiles,
	listAssistants,
	getFile,
	getKnowledgeFiles,
	uploadKnowledge,
	deleteKnowledgeFile,
	deleteFile,
	invoke,
	newMessageEventSource
} from '$lib/services/chat/operations';
import { baseURL } from '$lib/services/chat/http';

export default {
	progressToMessages: buildMessagesFromProgress,
	listFiles,
	listAssistants,
	getKnowledgeFiles,
	uploadKnowledge,
	deleteKnowledgeFile,
	deleteFile,
	getFile,
	invoke,
	newMessageEventSource,
	getProfile,
	baseURL
};
