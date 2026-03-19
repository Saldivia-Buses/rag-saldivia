import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import CollectionCard from './CollectionCard.svelte';
import type { CollectionStats } from '$lib/server/gateway';

function makeStats(overrides: Partial<CollectionStats> = {}): CollectionStats {
    return {
        collection: 'test-col',
        entity_count: 1234,
        index_type: 'HNSW',
        has_sparse: false,
        ...overrides,
    };
}

describe('CollectionCard', () => {
    it('renderiza el nombre de la colección', () => {
        render(CollectionCard, {
            props: { name: 'documentos-legales', stats: makeStats(), href: '/col/docs' },
        });
        expect(screen.getByText('documentos-legales')).toBeInTheDocument();
    });

    it('es un link con el href correcto', () => {
        render(CollectionCard, {
            props: { name: 'mi-col', stats: makeStats(), href: '/collections/mi-col' },
        });
        const link = screen.getByRole('link');
        expect(link).toHaveAttribute('href', '/collections/mi-col');
    });

    it('muestra entity_count formateado con stats presentes', () => {
        render(CollectionCard, {
            props: { name: 'col', stats: makeStats({ entity_count: 5678 }), href: '/c' },
        });
        // toLocaleString puede generar "5.678" o "5,678" según locale
        expect(screen.getByText(/5[\.,]?678/)).toBeInTheDocument();
    });

    it('muestra el index_type cuando está presente', () => {
        render(CollectionCard, {
            props: { name: 'col', stats: makeStats({ index_type: 'IVF_FLAT' }), href: '/c' },
        });
        expect(screen.getByText(/IVF_FLAT/)).toBeInTheDocument();
    });

    it('muestra badge Sparse cuando has_sparse=true', () => {
        render(CollectionCard, {
            props: { name: 'col', stats: makeStats({ has_sparse: true }), href: '/c' },
        });
        expect(screen.getByText('Sparse')).toBeInTheDocument();
    });

    it('NO muestra badge Sparse cuando has_sparse=false', () => {
        render(CollectionCard, {
            props: { name: 'col', stats: makeStats({ has_sparse: false }), href: '/c' },
        });
        expect(screen.queryByText('Sparse')).toBeNull();
    });

    it('muestra skeleton cuando stats es null', () => {
        const { container } = render(CollectionCard, {
            props: { name: 'cargando', stats: null, href: '/c' },
        });
        // Skeleton renderiza divs con clases de shimmer — el nombre sí se ve
        expect(screen.getByText('cargando')).toBeInTheDocument();
        // No debe mostrar entity_count (ya que no hay stats)
        expect(screen.queryByText(/entidades/)).toBeNull();
    });

    it('muestra texto "entidades" cuando hay stats', () => {
        render(CollectionCard, {
            props: { name: 'col', stats: makeStats(), href: '/c' },
        });
        expect(screen.getByText(/entidades/)).toBeInTheDocument();
    });
});
