import { buildMessagesFromProgress } from './messages';
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
	newMessageEventSource,
	listTools,
	enableTool,
	disableTool
} from './operations';
import { baseURL } from './http';

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
	listTools,
	enableTool,
	disableTool,
	baseURL
};
