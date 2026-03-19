import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { crossdoc } from '$lib/stores/crossdoc.svelte';
import CrossdocProgress from './CrossdocProgress.svelte';

beforeEach(() => {
    crossdoc.reset(); // progress = null entre tests
});

describe('CrossdocProgress', () => {
    it('no renderiza nada cuando progress es null', () => {
        const { container } = render(CrossdocProgress);
        // El #if p && p.phase !== 'done' && p.phase !== 'error' es falso
        expect(container.querySelector('.space-y-2')).toBeNull();
    });

    it('no renderiza nada cuando phase=done', () => {
        crossdoc.progress = { phase: 'done', subQueries: [], completed: 0, total: 0, results: [] };
        const { container } = render(CrossdocProgress);
        expect(container.querySelector('.space-y-2')).toBeNull();
    });

    it('no renderiza nada cuando phase=error', () => {
        crossdoc.progress = {
            phase: 'error' as any,
            subQueries: [], completed: 0, total: 0, results: [],
            error: 'algo salió mal',
        };
        const { container } = render(CrossdocProgress);
        expect(container.querySelector('.space-y-2')).toBeNull();
    });

    it('muestra las cuatro pills de fases cuando phase=decomposing', () => {
        crossdoc.progress = {
            phase: 'decomposing',
            subQueries: [], completed: 0, total: 0, results: [],
        };
        render(CrossdocProgress);

        // Cada pill renderiza "{icon} {label}" dentro del mismo span,
        // por eso se busca con regex parcial (no exact match)
        expect(screen.getByText(/Analizando pregunta/)).toBeInTheDocument();
        expect(screen.getByText(/Consultando documentos/)).toBeInTheDocument();
        expect(screen.getByText(/Reintentando fallidos/)).toBeInTheDocument();
        expect(screen.getByText(/Sintetizando respuesta/)).toBeInTheDocument();
    });

    it('muestra barra de progreso cuando phase=querying con total > 0', () => {
        crossdoc.progress = {
            phase: 'querying',
            subQueries: ['q1', 'q2'],
            completed: 1,
            total: 2,
            results: [],
        };
        render(CrossdocProgress);

        expect(screen.getByText('1 / 2 sub-queries')).toBeInTheDocument();
    });

    it('muestra barra de progreso cuando phase=retrying con total > 0', () => {
        crossdoc.progress = {
            phase: 'retrying',
            subQueries: ['q1'],
            completed: 0,
            total: 1,
            results: [],
        };
        render(CrossdocProgress);

        expect(screen.getByText('0 / 1 sub-queries')).toBeInTheDocument();
    });

    it('NO muestra barra de progreso cuando phase=synthesizing', () => {
        crossdoc.progress = {
            phase: 'synthesizing',
            subQueries: ['q1'],
            completed: 1,
            total: 1,
            results: [],
        };
        const { container } = render(CrossdocProgress);

        // synthesizing no tiene barra numérica (solo el bloque de fases)
        expect(screen.queryByText(/sub-queries/)).toBeNull();
        expect(container.querySelector('.space-y-2')).toBeInTheDocument();
    });
});
