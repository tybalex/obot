import queryString from "query-string";

import { ConsumptionUrl } from "~/lib/routers/baseRouter";

const buildUrl = (path: string, params?: Record<string, string>) => {
	const url = new URL(ConsumptionUrl(path));

	if (params) {
		url.search = queryString.stringify(params, { skipNull: true });
	}

	return {
		url: url.toString(),
		path: url.pathname,
	};
};

export const UserRoutes = {
	root: () => buildUrl("/"),
	obot: (projectId: string) => buildUrl(`/o/${projectId}`),
	thread: (projectId: string, threadId: string) =>
		buildUrl(`/o/${projectId}/t/${threadId}`),
};
