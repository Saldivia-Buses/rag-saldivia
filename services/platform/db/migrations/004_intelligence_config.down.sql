DELETE FROM agent_config WHERE updated_by = 'system';
DELETE FROM prompt_versions WHERE created_by = 'system';
DELETE FROM llm_models WHERE id IN ('paddleocr-vl', 'qwen3.5-9b');
DROP TABLE IF EXISTS trace_events;
DROP TABLE IF EXISTS execution_traces;
DROP TABLE IF EXISTS prompt_versions;
DROP TABLE IF EXISTS tool_registry;
DROP TABLE IF EXISTS llm_models;
DROP TABLE IF EXISTS agent_config;
