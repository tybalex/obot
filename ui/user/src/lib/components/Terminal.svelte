<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import '@xterm/xterm/css/xterm.css';
	import { currentAssistant } from '$lib/stores';
	import { RefreshCcw } from 'lucide-svelte';

	let terminalContainer: HTMLElement;
	let close: () => void;
	let connectState = $state('disconnected');

	onDestroy(() => close?.());

	onMount(connect);

	async function connect() {
		const { Terminal } = await import('@xterm/xterm');
		const { FitAddon } = await import('@xterm/addon-fit');

		close?.();

		const term = new Terminal();
		const fitAddon = new FitAddon();
		term.loadAddon(fitAddon);

		term.open(terminalContainer);

		new ResizeObserver(() => {
			fitAddon.fit();
		}).observe(terminalContainer);

		const url =
			window.location.protocol.replaceAll('http', 'ws') +
			'//' +
			window.location.host +
			'/api/assistants/' +
			$currentAssistant.id +
			'/shell';
		let gotData = false;
		const socket = new WebSocket(url);
		connectState = 'connecting';
		socket.onmessage = (event) => term.write(event.data);
		socket.onopen = () => {
			connectState = 'connected';
			fitAddon.fit();
			term.focus();
			setTimeout(() => {
				if (!gotData) {
					socket.send('\n');
				}
			}, 500);
		};
		socket.onclose = () => {
			connectState = 'disconnected';
			term.write('\r\nConnection closed.\r\n');
		};
		socket.onerror = () => {
			connectState = 'disconnected';
			term.write('\r\nConnection error.\r\n');
		};
		term.options.theme = {
			background: '#131313'
		};
		term.onData((data) => {
			gotData = true;
			socket.send(data);
		});
		term.onResize(({ cols, rows }) => {
			const data = JSON.stringify({ cols, rows });
			socket.send(new Blob([data], { type: 'application/json' }));
		});

		close = () => {
			socket.close();
			term.dispose();
		};
	}
</script>

<div class="flex h-full w-full flex-col">
	{#if connectState !== 'connected'}
		<div class="flex items-center gap-2 self-end">
			<span class="capitalize">{connectState}</span>
			{#if connectState === 'disconnected'}
				<button class="icon-button" onclick={connect}>
					<RefreshCcw class="h-4 w-4" />
				</button>
			{/if}
		</div>
	{/if}
	<div class="flex-1 rounded-3xl bg-gray-950 p-5">
		<div bind:this={terminalContainer}></div>
	</div>
</div>

<style lang="postcss">
	:global {
		.xterm > div {
			@apply scrollbar-none;
		}
	}
</style>
