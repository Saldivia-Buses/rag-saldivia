CREATE TABLE `annotations` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`user_id` integer NOT NULL,
	`session_id` text NOT NULL,
	`message_id` integer,
	`selected_text` text NOT NULL,
	`note` text,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`session_id`) REFERENCES `chat_sessions`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`message_id`) REFERENCES `chat_messages`(`id`) ON UPDATE no action ON DELETE set null
);
--> statement-breakpoint
CREATE INDEX `idx_annotations_user` ON `annotations` (`user_id`);--> statement-breakpoint
CREATE INDEX `idx_annotations_session` ON `annotations` (`session_id`);--> statement-breakpoint
CREATE TABLE `area_collections` (
	`area_id` integer NOT NULL,
	`collection_name` text NOT NULL,
	`permission` text DEFAULT 'read' NOT NULL,
	PRIMARY KEY(`area_id`, `collection_name`),
	FOREIGN KEY (`area_id`) REFERENCES `areas`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `areas` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`name` text NOT NULL,
	`description` text DEFAULT '' NOT NULL,
	`created_at` integer NOT NULL
);
--> statement-breakpoint
CREATE UNIQUE INDEX `areas_name_unique` ON `areas` (`name`);--> statement-breakpoint
CREATE TABLE `audit_log` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`user_id` integer NOT NULL,
	`action` text NOT NULL,
	`collection` text,
	`query_preview` text,
	`ip_address` text DEFAULT '' NOT NULL,
	`timestamp` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE INDEX `idx_audit_user` ON `audit_log` (`user_id`);--> statement-breakpoint
CREATE INDEX `idx_audit_timestamp` ON `audit_log` (`timestamp`);--> statement-breakpoint
CREATE TABLE `bot_user_mappings` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`platform` text NOT NULL,
	`external_user_id` text NOT NULL,
	`system_user_id` integer NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`system_user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `idx_bot_user_mapping_unique` ON `bot_user_mappings` (`platform`,`external_user_id`);--> statement-breakpoint
CREATE TABLE `chat_messages` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`session_id` text NOT NULL,
	`role` text NOT NULL,
	`content` text NOT NULL,
	`sources` text,
	`timestamp` integer NOT NULL,
	FOREIGN KEY (`session_id`) REFERENCES `chat_sessions`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_chat_messages_session` ON `chat_messages` (`session_id`);--> statement-breakpoint
CREATE TABLE `chat_sessions` (
	`id` text PRIMARY KEY NOT NULL,
	`user_id` integer NOT NULL,
	`title` text NOT NULL,
	`collection` text NOT NULL,
	`crossdoc` integer DEFAULT false NOT NULL,
	`forked_from` text,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_chat_sessions_user` ON `chat_sessions` (`user_id`);--> statement-breakpoint
CREATE INDEX `idx_chat_sessions_user_updated` ON `chat_sessions` (`user_id`,`updated_at`);--> statement-breakpoint
CREATE TABLE `collection_history` (
	`id` text PRIMARY KEY NOT NULL,
	`collection` text NOT NULL,
	`user_id` integer NOT NULL,
	`action` text NOT NULL,
	`filename` text,
	`doc_count` integer,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_collection_history_collection` ON `collection_history` (`collection`);--> statement-breakpoint
CREATE TABLE `events` (
	`id` text PRIMARY KEY NOT NULL,
	`ts` integer NOT NULL,
	`source` text NOT NULL,
	`level` text NOT NULL,
	`type` text NOT NULL,
	`user_id` integer,
	`session_id` text,
	`payload` text DEFAULT '{}' NOT NULL,
	`sequence` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE INDEX `idx_events_ts` ON `events` (`ts`);--> statement-breakpoint
CREATE INDEX `idx_events_type` ON `events` (`type`);--> statement-breakpoint
CREATE INDEX `idx_events_user` ON `events` (`user_id`);--> statement-breakpoint
CREATE INDEX `idx_events_level` ON `events` (`level`);--> statement-breakpoint
CREATE INDEX `idx_events_sequence` ON `events` (`sequence`);--> statement-breakpoint
CREATE TABLE `external_sources` (
	`id` text PRIMARY KEY NOT NULL,
	`user_id` integer NOT NULL,
	`provider` text NOT NULL,
	`name` text NOT NULL,
	`credentials` text DEFAULT '{}' NOT NULL,
	`collection_dest` text NOT NULL,
	`schedule` text DEFAULT 'daily' NOT NULL,
	`active` integer DEFAULT true NOT NULL,
	`last_sync` integer,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_external_sources_user` ON `external_sources` (`user_id`);--> statement-breakpoint
CREATE TABLE `ingestion_alerts` (
	`id` text PRIMARY KEY NOT NULL,
	`job_id` text NOT NULL,
	`user_id` integer NOT NULL,
	`filename` text NOT NULL,
	`collection` text NOT NULL,
	`tier` text NOT NULL,
	`page_count` integer,
	`file_hash` text,
	`error` text,
	`retry_count` integer,
	`progress_at_failure` integer,
	`gateway_version` text,
	`created_at` integer NOT NULL,
	`resolved_at` integer,
	`resolved_by` text,
	`notes` text,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE INDEX `idx_alerts_resolved` ON `ingestion_alerts` (`resolved_at`);--> statement-breakpoint
CREATE TABLE `ingestion_jobs` (
	`id` text PRIMARY KEY NOT NULL,
	`user_id` integer NOT NULL,
	`task_id` text NOT NULL,
	`filename` text NOT NULL,
	`collection` text NOT NULL,
	`tier` text NOT NULL,
	`page_count` integer,
	`state` text DEFAULT 'pending' NOT NULL,
	`progress` integer DEFAULT 0 NOT NULL,
	`file_hash` text,
	`retry_count` integer DEFAULT 0 NOT NULL,
	`last_checked` integer,
	`created_at` integer NOT NULL,
	`completed_at` integer,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE INDEX `idx_ingestion_jobs_user` ON `ingestion_jobs` (`user_id`);--> statement-breakpoint
CREATE INDEX `idx_ingestion_jobs_state` ON `ingestion_jobs` (`state`);--> statement-breakpoint
CREATE TABLE `ingestion_queue` (
	`id` text PRIMARY KEY NOT NULL,
	`collection` text NOT NULL,
	`file_path` text NOT NULL,
	`user_id` integer NOT NULL,
	`priority` integer DEFAULT 0 NOT NULL,
	`status` text DEFAULT 'pending' NOT NULL,
	`locked_at` integer,
	`locked_by` text,
	`created_at` integer NOT NULL,
	`started_at` integer,
	`completed_at` integer,
	`error` text,
	`retry_count` integer DEFAULT 0 NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE INDEX `idx_queue_status` ON `ingestion_queue` (`status`,`priority`);--> statement-breakpoint
CREATE INDEX `idx_queue_pending` ON `ingestion_queue` (`status`,`locked_at`);--> statement-breakpoint
CREATE TABLE `message_feedback` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`message_id` integer NOT NULL,
	`user_id` integer NOT NULL,
	`rating` text NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`message_id`) REFERENCES `chat_messages`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE UNIQUE INDEX `idx_feedback_unique` ON `message_feedback` (`message_id`,`user_id`);--> statement-breakpoint
CREATE INDEX `idx_message_feedback_message` ON `message_feedback` (`message_id`);--> statement-breakpoint
CREATE TABLE `project_collections` (
	`project_id` text NOT NULL,
	`collection_name` text NOT NULL,
	PRIMARY KEY(`project_id`, `collection_name`),
	FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `project_sessions` (
	`project_id` text NOT NULL,
	`session_id` text NOT NULL,
	PRIMARY KEY(`project_id`, `session_id`),
	FOREIGN KEY (`project_id`) REFERENCES `projects`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`session_id`) REFERENCES `chat_sessions`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `projects` (
	`id` text PRIMARY KEY NOT NULL,
	`user_id` integer NOT NULL,
	`name` text NOT NULL,
	`description` text DEFAULT '' NOT NULL,
	`instructions` text DEFAULT '' NOT NULL,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_projects_user` ON `projects` (`user_id`);--> statement-breakpoint
CREATE TABLE `prompt_templates` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`title` text NOT NULL,
	`prompt` text NOT NULL,
	`focus_mode` text DEFAULT 'detallado' NOT NULL,
	`created_by` integer NOT NULL,
	`active` integer DEFAULT true NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`created_by`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_prompt_templates_active` ON `prompt_templates` (`active`);--> statement-breakpoint
CREATE TABLE `rate_limits` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`target_type` text NOT NULL,
	`target_id` integer NOT NULL,
	`max_queries_per_hour` integer NOT NULL,
	`active` integer DEFAULT true NOT NULL,
	`created_at` integer NOT NULL
);
--> statement-breakpoint
CREATE INDEX `idx_rate_limits_target` ON `rate_limits` (`target_type`,`target_id`);--> statement-breakpoint
CREATE TABLE `saved_responses` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`user_id` integer NOT NULL,
	`message_id` integer,
	`content` text NOT NULL,
	`session_title` text,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`message_id`) REFERENCES `chat_messages`(`id`) ON UPDATE no action ON DELETE set null
);
--> statement-breakpoint
CREATE INDEX `idx_saved_responses_user` ON `saved_responses` (`user_id`);--> statement-breakpoint
CREATE TABLE `scheduled_reports` (
	`id` text PRIMARY KEY NOT NULL,
	`user_id` integer NOT NULL,
	`query` text NOT NULL,
	`collection` text NOT NULL,
	`schedule` text NOT NULL,
	`destination` text NOT NULL,
	`email` text,
	`active` integer DEFAULT true NOT NULL,
	`last_run` integer,
	`next_run` integer NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_reports_active_next_run` ON `scheduled_reports` (`active`,`next_run`);--> statement-breakpoint
CREATE TABLE `session_shares` (
	`id` text PRIMARY KEY NOT NULL,
	`session_id` text NOT NULL,
	`user_id` integer NOT NULL,
	`token` text NOT NULL,
	`expires_at` integer NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`session_id`) REFERENCES `chat_sessions`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `session_shares_token_unique` ON `session_shares` (`token`);--> statement-breakpoint
CREATE UNIQUE INDEX `idx_session_shares_token` ON `session_shares` (`token`);--> statement-breakpoint
CREATE TABLE `session_tags` (
	`session_id` text NOT NULL,
	`tag` text NOT NULL,
	PRIMARY KEY(`session_id`, `tag`),
	FOREIGN KEY (`session_id`) REFERENCES `chat_sessions`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_session_tags_tag` ON `session_tags` (`tag`);--> statement-breakpoint
CREATE TABLE `user_areas` (
	`user_id` integer NOT NULL,
	`area_id` integer NOT NULL,
	PRIMARY KEY(`user_id`, `area_id`),
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`area_id`) REFERENCES `areas`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `user_memory` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`user_id` integer NOT NULL,
	`key` text NOT NULL,
	`value` text NOT NULL,
	`source` text DEFAULT 'explicit' NOT NULL,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `idx_user_memory_unique` ON `user_memory` (`user_id`,`key`);--> statement-breakpoint
CREATE TABLE `users` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`email` text NOT NULL,
	`name` text NOT NULL,
	`role` text DEFAULT 'user' NOT NULL,
	`api_key_hash` text NOT NULL,
	`password_hash` text,
	`preferences` text DEFAULT '{}' NOT NULL,
	`active` integer DEFAULT true NOT NULL,
	`onboarding_completed` integer DEFAULT false NOT NULL,
	`sso_provider` text,
	`sso_subject` text,
	`created_at` integer NOT NULL,
	`last_login` integer
);
--> statement-breakpoint
CREATE UNIQUE INDEX `users_email_unique` ON `users` (`email`);--> statement-breakpoint
CREATE INDEX `idx_users_api_key` ON `users` (`api_key_hash`);--> statement-breakpoint
CREATE TABLE `webhooks` (
	`id` text PRIMARY KEY NOT NULL,
	`user_id` integer NOT NULL,
	`url` text NOT NULL,
	`events` text DEFAULT '[]' NOT NULL,
	`secret` text NOT NULL,
	`active` integer DEFAULT true NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE INDEX `idx_webhooks_active` ON `webhooks` (`active`);--> statement-breakpoint
-- FTS5 virtual tables para búsqueda universal (F3.39)
-- Drizzle Kit no genera virtual tables — se incluyen manualmente
CREATE VIRTUAL TABLE IF NOT EXISTS sessions_fts USING fts5(
  session_id UNINDEXED,
  user_id UNINDEXED,
  title,
  collection UNINDEXED,
  content=chat_sessions,
  content_rowid=rowid
);--> statement-breakpoint
CREATE VIRTUAL TABLE IF NOT EXISTS messages_fts USING fts5(
  message_id UNINDEXED,
  session_id UNINDEXED,
  body,
  content=chat_messages,
  content_rowid=rowid
);--> statement-breakpoint
-- Triggers para mantener sessions_fts sincronizado
CREATE TRIGGER IF NOT EXISTS sessions_fts_insert AFTER INSERT ON chat_sessions BEGIN
  INSERT INTO sessions_fts(rowid, session_id, user_id, title, collection) VALUES (new.rowid, new.id, new.user_id, new.title, new.collection);
END;--> statement-breakpoint
CREATE TRIGGER IF NOT EXISTS sessions_fts_delete AFTER DELETE ON chat_sessions BEGIN
  INSERT INTO sessions_fts(sessions_fts, rowid, session_id, user_id, title, collection) VALUES ('delete', old.rowid, old.id, old.user_id, old.title, old.collection);
END;--> statement-breakpoint
CREATE TRIGGER IF NOT EXISTS sessions_fts_update AFTER UPDATE ON chat_sessions BEGIN
  INSERT INTO sessions_fts(sessions_fts, rowid, session_id, user_id, title, collection) VALUES ('delete', old.rowid, old.id, old.user_id, old.title, old.collection);
  INSERT INTO sessions_fts(rowid, session_id, user_id, title, collection) VALUES (new.rowid, new.id, new.user_id, new.title, new.collection);
END;--> statement-breakpoint
-- Triggers para mantener messages_fts sincronizado
CREATE TRIGGER IF NOT EXISTS messages_fts_insert AFTER INSERT ON chat_messages BEGIN
  INSERT INTO messages_fts(rowid, message_id, session_id, body) VALUES (new.rowid, new.id, new.session_id, new.content);
END;--> statement-breakpoint
CREATE TRIGGER IF NOT EXISTS messages_fts_delete AFTER DELETE ON chat_messages BEGIN
  INSERT INTO messages_fts(messages_fts, rowid, message_id, session_id, body) VALUES ('delete', old.rowid, old.id, old.session_id, old.content);
END;