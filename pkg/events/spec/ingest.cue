package events

// Ingest family — events emitted by the ingestion pipeline.

events: "ingest.completed": {
	version: 1
	subject: "tenant.{slug}.ingest.completed"
	payload: {
		job_id:          string
		collection_name: string
		doc_count:       int
		chunk_count:     int
		duration_ms:     int
	}
	publishers: ["ingest"]
	consumers:  ["notification", "ws"]
}
