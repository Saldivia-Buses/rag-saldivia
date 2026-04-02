import type { Action } from 'svelte/action';

export const clickOutside: Action<HTMLElement, () => void> = (node, callback) => {
	function handle(event: MouseEvent) {
		if (!node.contains(event.target as Node)) {
			callback?.();
		}
	}
	document.addEventListener('click', handle, true);
	return {
		destroy() {
			document.removeEventListener('click', handle, true);
		},
		update(newCallback) {
			callback = newCallback;
		},
	};
};
