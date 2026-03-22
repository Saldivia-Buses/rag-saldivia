<!-- services/sda-frontend/src/lib/components/layout/CommandPalette.svelte -->
<!-- TODO Fase 5: implementar ⌘K command palette con búsqueda de rutas y acciones -->
<script lang="ts">
    interface Props {
        open?: boolean;
    }
    let { open = $bindable(false) }: Props = $props();
</script>

<!-- Stub: CommandPalette se implementa en Fase 5 -->
<!-- open={open} se expondrá para control externo cuando se implemente -->
