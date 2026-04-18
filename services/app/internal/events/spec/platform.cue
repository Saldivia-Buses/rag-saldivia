package events

// Platform family — cross-tenant lifecycle events (NOT tenant-namespaced).
// Consumed by services that need to react to tenant onboarding/offboarding
// (DrainerRegistry hotload, auth cache invalidation, etc.).

events: "platform.lifecycle": {
	version: 1
	subject: "platform.lifecycle.{action}"
	payload: {
		action:      "tenant_created" | "tenant_deleted" | "tenant_suspended"
		tenant_id:   string
		tenant_slug: string
		by_user_id:  string
	}
	publishers: ["platform"]
	consumers:  ["auth", "chat", "ingest", "healthwatch", "ws"]
}
