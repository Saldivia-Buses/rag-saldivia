<script lang="ts">
    import { enhance } from '$app/forms';
    import Input from '$lib/components/ui/Input.svelte';
    import Button from '$lib/components/ui/Button.svelte';

    let { form } = $props();
    let submitting = $state(false);
</script>

<div class="min-h-screen flex items-center justify-center bg-[var(--bg-base)] p-6">
    <div class="w-full max-w-sm">
        <!-- Logo -->
        <div class="text-center mb-8">
            <div class="inline-flex items-center justify-center w-12 h-12 bg-[var(--accent)] rounded-2xl mb-3">
                <span class="text-white font-bold text-lg">S</span>
            </div>
            <h1 class="text-lg font-bold text-[var(--text)]">SDA</h1>
            <p class="text-sm text-[var(--text-muted)] mt-1">Sistema de Documentación Asistida</p>
            <p class="text-xs text-[var(--text-faint)] mt-0.5">Saldivia Buses</p>
        </div>

        <!-- Form -->
        <div class="bg-[var(--bg-surface)] border border-[var(--border)] rounded-[var(--radius-lg)] p-6">
            {#if form?.error}
                <div class="bg-[var(--danger-bg)] border border-[var(--danger)] rounded-[var(--radius-md)] px-3 py-2 text-sm text-[var(--danger)] mb-4">
                    {form.error}
                </div>
            {/if}

            <form method="POST" use:enhance={() => {
                submitting = true;
                return async ({ update }) => { submitting = false; await update(); };
            }} class="flex flex-col gap-4">
                <Input
                    name="email"
                    type="email"
                    label="Email"
                    placeholder="usuario@saldivia.com.ar"
                    required
                    disabled={submitting}
                />
                <Input
                    name="password"
                    type="password"
                    label="Contraseña"
                    required
                    disabled={submitting}
                />
                <!-- Wrapper makes the inline-flex button stretch to full width -->
                <div class="mt-2 flex flex-col">
                    <Button type="submit" size="lg" loading={submitting}>
                        Ingresar
                    </Button>
                </div>
            </form>
        </div>
    </div>
</div>
