DROP TABLE `ingestion_queue`;--> statement-breakpoint
CREATE INDEX `idx_events_query` ON `events` (`type`,`user_id`,`ts`);