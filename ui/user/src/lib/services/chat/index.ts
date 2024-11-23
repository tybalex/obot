import { baseURL } from './http';
import { buildMessagesFromProgress } from './messages';
import * as MessageSource from './messagesource';
import * as Operations from './operations';

export default {
	progressToMessages: buildMessagesFromProgress,
	baseURL,
	...Operations,
	...MessageSource
};
