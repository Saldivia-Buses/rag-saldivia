// @vitest-environment jsdom
import { describe, it, expect, vi, afterEach } from 'vitest';
import { clickOutside } from './clickOutside.js';

describe('clickOutside action', () => {
	afterEach(() => {
		document.body.innerHTML = '';
	});

	it('calls callback when clicking outside the element', () => {
		const node = document.createElement('div');
		const outside = document.createElement('button');
		document.body.appendChild(node);
		document.body.appendChild(outside);
		const handler = vi.fn();

		clickOutside(node, handler);
		outside.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(handler).toHaveBeenCalledOnce();
	});

	it('does not call callback when clicking inside the element', () => {
		const node = document.createElement('div');
		const child = document.createElement('span');
		node.appendChild(child);
		document.body.appendChild(node);
		const handler = vi.fn();

		clickOutside(node, handler);
		child.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(handler).not.toHaveBeenCalled();
	});

	it('does not call callback when clicking the node itself', () => {
		const node = document.createElement('div');
		document.body.appendChild(node);
		const handler = vi.fn();

		clickOutside(node, handler);
		node.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(handler).not.toHaveBeenCalled();
	});

	it('cleans up listener on destroy', () => {
		const node = document.createElement('div');
		const outside = document.createElement('button');
		document.body.appendChild(node);
		document.body.appendChild(outside);
		const handler = vi.fn();

		const action = clickOutside(node, handler);
		action?.destroy?.();
		outside.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(handler).not.toHaveBeenCalled();
	});

	it('updates callback when update is called', () => {
		const node = document.createElement('div');
		const outside = document.createElement('button');
		document.body.appendChild(node);
		document.body.appendChild(outside);
		const handler1 = vi.fn();
		const handler2 = vi.fn();

		const action = clickOutside(node, handler1);
		action?.update?.(handler2);
		outside.dispatchEvent(new MouseEvent('click', { bubbles: true }));

		expect(handler1).not.toHaveBeenCalled();
		expect(handler2).toHaveBeenCalledOnce();
	});
});
