interface ScrollMetrics {
    scrollHeight: number;
    scrollTop: number;
    clientHeight: number;
}

/**
 * Retorna true si el elemento está dentro de `threshold` píxeles del fondo.
 * Útil para decidir si auto-scrollear o mostrar el botón "↓ Ir al fondo".
 */
export function isNearBottom(el: ScrollMetrics, threshold = 100): boolean {
    return el.scrollHeight - el.scrollTop - el.clientHeight < threshold;
}
