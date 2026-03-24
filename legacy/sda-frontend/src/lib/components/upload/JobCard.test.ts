import { describe, it, expect, vi } from 'vitest';
import { render } from '@testing-library/svelte';
import JobCard from './JobCard.svelte';

const baseJob = {
    jobId: 'j1', filename: 'contrato.pdf', collection: 'legal',
    tier: 'medium' as const, pageCount: 180,
    state: 'running' as const, progress: 75,
    eta: 30, startedAt: Date.now(), lastProgressAt: Date.now(),
};

describe('JobCard', () => {
    it('muestra el nombre del archivo', () => {
        const { getByText } = render(JobCard, { props: { job: baseJob } });
        expect(getByText('contrato.pdf')).toBeTruthy();
    });

    it('barra de progreso al 75%', () => {
        const { container } = render(JobCard, { props: { job: baseJob } });
        const bar = container.querySelector('[style*="width: 75%"]');
        expect(bar).toBeTruthy();
    });

    it('muestra botón reintentar cuando state=stalled', () => {
        const stalledJob = { ...baseJob, state: 'stalled' as const };
        const onRetry = vi.fn();
        const { getByText } = render(JobCard, { props: { job: stalledJob, onRetry } });
        expect(getByText('Reintentar')).toBeTruthy();
    });

    it('no muestra botón reintentar cuando state=running', () => {
        const { queryByText } = render(JobCard, { props: { job: baseJob } });
        expect(queryByText('Reintentar')).toBeNull();
    });

    it('muestra checkmark cuando state=completed', () => {
        const done = { ...baseJob, state: 'completed' as const, progress: 100 };
        const { container } = render(JobCard, { props: { job: done } });
        expect(container.querySelector('svg')).toBeTruthy();
    });
});
