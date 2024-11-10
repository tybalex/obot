import { buildMessagesFromProgress } from './messages';
import {
	getProfile,
	listFiles,
	listAssistants,
	getFile,
	listKnowledgeFiles,
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
	listKnowledgeFiles,
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
