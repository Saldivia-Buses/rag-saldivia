<script lang="ts">
    import Sidebar from '$lib/components/layout/Sidebar.svelte';
    import ToastContainer from '$lib/components/ui/ToastContainer.svelte';

    interface Props { data: any; children: any; }
    let { data, children }: Props = $props();
</script>

<div class="flex h-screen overflow-hidden bg-[var(--bg-base)]">
    <Sidebar
        role={data.user.role}
        userName={data.user.name ?? data.user.email}
        userEmail={data.user.email}
    />

    <main class="flex-1 overflow-auto min-w-0">
        {@render children()}
    </main>
</div>

<ToastContainer />
