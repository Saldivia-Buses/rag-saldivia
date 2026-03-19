export type ToastType = 'success' | 'error' | 'info' | 'warning';

export interface Toast {
    id: string;
    type: ToastType;
    message: string;
    duration: number;
}

class ToastStore {
    toasts = $state<Toast[]>([]);

    private add(type: ToastType, message: string, duration: number) {
        const id = crypto.randomUUID();
        this.toasts.push({ id, type, message, duration });
        if (duration > 0) {
            setTimeout(() => this.dismiss(id), duration);
        }
    }

    success(message: string, duration = 4000) { this.add('success', message, duration); }
    error(message: string, duration = 6000)   { this.add('error',   message, duration); }
    info(message: string, duration = 4000)    { this.add('info',    message, duration); }
    warning(message: string, duration = 5000) { this.add('warning', message, duration); }

    dismiss(id: string) {
        this.toasts = this.toasts.filter(t => t.id !== id);
    }
}

export const toastStore = new ToastStore();
