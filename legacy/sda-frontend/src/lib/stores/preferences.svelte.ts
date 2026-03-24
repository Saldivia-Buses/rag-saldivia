import type { UserPreferences } from '$lib/types/preferences';
import { DEFAULT_PREFERENCES } from '$lib/types/preferences';

export class PreferencesStore {
    current = $state<UserPreferences>({ ...DEFAULT_PREFERENCES });

    init(prefs: UserPreferences) {
        this.current = prefs;
    }

    get avatarColor() { return this.current.avatar_color; }
    get language() { return this.current.ui_language; }
}

export const preferencesStore = new PreferencesStore();
