package events

// Notify family — events consumed by the notification service (and forwarded
// to the WS hub for in-app badges). See docs/conventions/cue.md for breaking
// change rules.

events: "chat.new_message": {
	version: 1
	subject: "tenant.{slug}.notify.chat.new_message"
	payload: {
		user_id:    string // target user (who gets notified)
		session_id: string
		message_id: string
		title:      string
		body:       string
		channel:    "in_app" | "email" | "both"
	}
	publishers: ["chat"]
	consumers:  ["notification", "ws"]
}

events: "auth.login_success": {
	version: 1
	subject: "tenant.{slug}.auth.login_success"
	payload: {
		user_id:    string
		email:      string
		ip_address: string
		user_agent: string
	}
	publishers: ["auth"]
	consumers:  ["notification"]
}
