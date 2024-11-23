function resize(node: HTMLTextAreaElement) {
	node.style.height = 'auto';
	// not totally sure why 4 is needed here, otherwise the textarea is too small and we
	// get a scrollbar
	node.style.height = node.scrollHeight + 4 + 'px';
}

export function autoHeight(node: HTMLTextAreaElement) {
	resize(node);
	node.onkeydown = () => resize(node);
	node.onkeyup = () => resize(node);
	node.oninput = () => resize(node);
}
