import { tick } from 'svelte';

export function autoscroll(node: HTMLElement) {
	const observer = new MutationObserver(() => {
		if (node.dataset.autoscroll !== 'false') {
			tick().then(() => {
				node.scrollTop = node.scrollHeight;
			});
		}
	});

	observer.observe(node, { childList: true, subtree: true });

	node.addEventListener('scroll', () => {
		// Calculate the scroll position
		const scrollTop = node.scrollTop;
		const scrollHeight = node.scrollHeight;
		const clientHeight = node.clientHeight;

		// Check if the user has scrolled to withing 40px of the bottom
		if (scrollHeight - (scrollTop + clientHeight) <= 40) {
			node.dataset.autoscroll = 'true';
		} else {
			node.dataset.autoscroll = 'false';
		}
	});
}
