export interface UserPreferences {
    default_collection: string;
    default_query_mode: 'standard' | 'crossdoc';
    vdb_top_k: number;
    reranker_top_k: number;
    max_sub_queries: number;
    follow_up_retries: boolean;
    show_decomposition: boolean;
    avatar_color: string;
    ui_language: 'es' | 'en';
    notify_ingestion_done: boolean;
    notify_system_alerts: boolean;
}

export const DEFAULT_PREFERENCES: UserPreferences = {
    default_collection: '',
    default_query_mode: 'standard',
    vdb_top_k: 10,
    reranker_top_k: 5,
    max_sub_queries: 4,
    follow_up_retries: true,
    show_decomposition: false,
    avatar_color: '#6366f1',
    ui_language: 'es',
    notify_ingestion_done: true,
    notify_system_alerts: true,
};
