<script lang="ts">
    import { enhance } from '$app/forms';
    let { form } = $props();
    let submitting = $state(false);
</script>

<div class="min-h-screen flex items-center justify-center bg-[#070d1a]">
    <div class="w-full max-w-sm">
        <!-- Logo -->
        <div class="text-center mb-8">
            <div class="inline-block bg-[#6366f1] text-white font-bold text-lg px-4 py-2 rounded">
                SDA
            </div>
            <p class="text-[#475569] mt-2 text-sm">Sistema de Documentación Asistida</p>
        </div>

        <!-- Form -->
        <form method="POST" use:enhance={() => {
            submitting = true;
            return async ({ update }) => {
                submitting = false;
                await update();
            };
        }} class="bg-[#0f172a] border border-[#1e293b] rounded-lg p-8">
            {#if form?.error}
                <p class="text-red-400 text-sm mb-4">{form.error}</p>
            {/if}

            <div class="mb-4">
                <label for="email" class="block text-sm text-[#94a3b8] mb-1">Email</label>
                <input
                    id="email" name="email" type="email" required
                    disabled={submitting}
                    class="w-full bg-[#1e293b] border border-[#334155] rounded px-3 py-2
                           text-[#e2e8f0] text-sm focus:outline-none focus:border-[#6366f1]
                           disabled:opacity-50"
                />
            </div>

            <div class="mb-6">
                <label for="password" class="block text-sm text-[#94a3b8] mb-1">Contraseña</label>
                <input
                    id="password" name="password" type="password" required
                    disabled={submitting}
                    class="w-full bg-[#1e293b] border border-[#334155] rounded px-3 py-2
                           text-[#e2e8f0] text-sm focus:outline-none focus:border-[#6366f1]
                           disabled:opacity-50"
                />
            </div>

            <button
                type="submit"
                disabled={submitting}
                class="w-full bg-[#6366f1] hover:bg-[#4f46e5] text-white rounded py-2 text-sm
                       font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
                {submitting ? 'Ingresando...' : 'Ingresar'}
            </button>
        </form>
    </div>
</div>
