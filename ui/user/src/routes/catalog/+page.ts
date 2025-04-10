import { browser } from '$app/environment';
import { ChatService } from '$lib/services';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch }) => {
	try {
		const shares = ChatService.listProjectShares({ fetch });
		const tools = ChatService.listAllTools({ fetch });

		if (browser) {
			localStorage.setItem('hasVisitedCatalog', 'true');
		}

		return {
			shares: (await shares).items,
			tools: (await tools).items
		};
	} catch {
		return {
			shares: [],
			tools: []
		};
	}
};
