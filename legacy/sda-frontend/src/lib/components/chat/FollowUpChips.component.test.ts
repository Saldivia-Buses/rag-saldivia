// @vitest-environment jsdom
import { describe, it, expect, vi } from 'vitest';
import { render, fireEvent } from '@testing-library/svelte';
import FollowUpChips from './FollowUpChips.svelte';

describe('FollowUpChips', () => {
    it('renderiza un chip por sugerencia', () => {
        const { getAllByRole } = render(FollowUpChips, {
            props: { suggestions: ['¿Podés ampliar?', '¿Cuáles son los riesgos?', '¿Hay más docs?'], onselect: vi.fn() }
        });
        expect(getAllByRole('button')).toHaveLength(3);
    });

    it('click llama a onselect con el texto del chip', async () => {
        const onselect = vi.fn();
        const { getByText } = render(FollowUpChips, {
            props: { suggestions: ['¿Podés ampliar sobre el tema?'], onselect }
        });
        await fireEvent.click(getByText('¿Podés ampliar sobre el tema?'));
        expect(onselect).toHaveBeenCalledWith('¿Podés ampliar sobre el tema?');
    });

    it('no renderiza nada si suggestions está vacío', () => {
        const { container } = render(FollowUpChips, {
            props: { suggestions: [], onselect: vi.fn() }
        });
        expect(container.querySelectorAll('button')).toHaveLength(0);
    });
});
