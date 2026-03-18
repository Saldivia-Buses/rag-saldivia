<script lang="ts">
    import Sidebar from '$lib/components/sidebar/Sidebar.svelte';

    interface Props {
        data: any;
        children: any;
    }

    let { data, children }: Props = $props();
</script>

<div class="flex min-h-screen bg-[#0f172a]">
    <Sidebar role={data.user.role} areaId={data.user.area_id} />
    <main class="flex-1 overflow-auto">
        {@render children()}
    </main>
</div>
