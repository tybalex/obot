import {
	autoUpdate,
	computePosition,
	flip,
	shift,
	offset,
	type ComputePositionConfig
} from '@floating-ui/dom';
import type { Action, ActionReturn } from 'svelte/action';
import { type Readable, writable, type Writable } from 'svelte/store';

interface Popover extends Readable<boolean> {
	ref: Action;
	tooltip: Action;
	open: Writable<boolean>;
	toggle: () => void;
}

interface PopoverOptions extends Partial<ComputePositionConfig> {
	hover?: boolean;
}

export default function popover(opts?: PopoverOptions): Popover {
	let ref: HTMLElement;
	let tooltip: HTMLElement;
	const open = writable(false);

	function build(): ActionReturn | void {
		if (!ref || !tooltip) return;

		function updatePosition() {
			computePosition(ref, tooltip, {
				placement: 'bottom-end',
				middleware: [
					flip(),
					shift({
						padding: 8
					}),
					offset(8)
				],
				...opts
			}).then(({ x, y }) => {
				Object.assign(tooltip.style, {
					left: `${x}px`,
					top: `${y}px`
				});
			});
		}

		tooltip.classList.add('hidden');
		tooltip.classList.add('absolute');

		if (opts?.hover) {
			ref.addEventListener('mouseenter', () => {
				open.set(true);
			});
			ref.addEventListener('mouseleave', () => {
				open.set(false);
			});
		}

		let close: (() => void) | null;
		open.subscribe((value) => {
			if (value) {
				tooltip.classList.remove('hidden');
				updatePosition();
				close = autoUpdate(ref, tooltip, updatePosition);
			} else {
				if (close) {
					close();
				}
				tooltip.classList.add('hidden');
				close = null;
			}
		});

		return {
			destroy: function () {
				if (close) {
					close();
				}
			}
		};
	}

	return {
		ref: (node: HTMLElement) => {
			ref = node;
			return build();
		},
		tooltip: (node: HTMLElement) => {
			tooltip = node;
			return build();
		},
		open,
		subscribe: open.subscribe,
		toggle: () => {
			open.update((value) => !value);
		}
	};
}
