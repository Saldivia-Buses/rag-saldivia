package hub

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	chatservice "github.com/Camionerou/rag-saldivia/services/app/internal/realtime/chat/service"
)

// MutationHandler routes mutation messages to the chat service in-process.
//
// Pre-ADR-025-realtime this dialed a gRPC client against the standalone chat
// service. After the fusion, chat lives next door under internal/realtime/chat
// so we invoke its service methods directly — one fewer hop, one fewer client
// to configure, one fewer failure mode.
type MutationHandler struct {
	chat *chatservice.Chat
}

// NewMutationHandler creates a handler that dispatches mutations to chat.
// Returns nil if chat is nil (mutations stay disabled — hub tolerates this).
func NewMutationHandler(chat *chatservice.Chat) *MutationHandler {
	if chat == nil {
		return nil
	}
	return &MutationHandler{chat: chat}
}

// Close is a no-op — kept so callers can keep their shutdown hooks symmetrical.
// The old gRPC connection close lived here; there is nothing to close now.
func (h *MutationHandler) Close() {}

// Handle dispatches a mutation message to the chat service.
// Runs in a goroutine to avoid blocking the hub event loop.
func (h *MutationHandler) Handle(client *Client, msg Message) {
	go func() {
		result, err := h.dispatch(client, msg)
		if err != nil {
			client.SendMessage(Message{Type: Error, ID: msg.ID, Error: userFacingError(err)})
			return
		}
		client.SendMessage(Message{
			Type: Event,
			ID:   msg.ID,
			Data: result,
		})
	}()
}

func (h *MutationHandler) dispatch(client *Client, msg Message) (json.RawMessage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch msg.Action {
	case "create_session":
		var req struct {
			Title      string  `json:"title"`
			Collection *string `json:"collection,omitempty"`
		}
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		s, err := h.chat.CreateSession(ctx, client.UserID, req.Title, req.Collection)
		if err != nil {
			return nil, err
		}
		return json.Marshal(s)

	case "delete_session":
		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		if err := h.chat.DeleteSession(ctx, req.SessionID, client.UserID); err != nil {
			return nil, err
		}
		return json.RawMessage(`{}`), nil

	case "rename_session":
		var req struct {
			SessionID string `json:"session_id"`
			Title     string `json:"title"`
		}
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		if err := h.chat.RenameSession(ctx, req.SessionID, client.UserID, req.Title); err != nil {
			return nil, err
		}
		return json.RawMessage(`{}`), nil

	case "send_message":
		var req struct {
			SessionID string          `json:"session_id"`
			Role      string          `json:"role"`
			Content   string          `json:"content"`
			Thinking  *string         `json:"thinking,omitempty"`
			Sources   json.RawMessage `json:"sources,omitempty"`
			Metadata  json.RawMessage `json:"metadata,omitempty"`
		}
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		m, err := h.chat.AddMessage(ctx, req.SessionID, client.UserID,
			req.Role, req.Content, req.Thinking, req.Sources, req.Metadata)
		if err != nil {
			return nil, err
		}
		return json.Marshal(m)

	default:
		return nil, &unknownActionError{action: msg.Action}
	}
}

// userFacingError maps internal errors to short strings safe to ship to the
// browser. The old gRPC path inspected status.Code() and mapped Unauthenticated
// to a "token_expired" hint — the in-process path has no token check (WS
// already authenticated the client on upgrade), so that branch is gone.
func userFacingError(err error) string {
	switch {
	case errors.Is(err, chatservice.ErrSessionNotFound):
		return "session not found"
	case errors.Is(err, chatservice.ErrNotOwner):
		return "permission denied"
	}
	slog.Debug("mutation failed", "error", err)
	return "mutation failed"
}

type unknownActionError struct{ action string }

func (e *unknownActionError) Error() string { return "unknown mutation action: " + e.action }
