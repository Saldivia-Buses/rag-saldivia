package hub

import (
	"context"
	"encoding/json"
	"log/slog"

	chatv1 "github.com/Camionerou/rag-saldivia/gen/go/chat/v1"
	sdagrpc "github.com/Camionerou/rag-saldivia/pkg/grpc"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// MutationHandler routes mutation messages to the appropriate gRPC service.
type MutationHandler struct {
	chatClient chatv1.ChatServiceClient
	chatConn   *grpc.ClientConn
}

// NewMutationHandler creates a handler that routes mutations to Chat via gRPC.
// Returns nil if the chat gRPC target is empty (mutations stay as stub).
func NewMutationHandler(chatGRPCTarget string) *MutationHandler {
	if chatGRPCTarget == "" {
		return nil
	}

	conn, err := sdagrpc.Dial(chatGRPCTarget)
	if err != nil {
		slog.Warn("chat grpc client failed, mutations disabled", "error", err)
		return nil
	}

	slog.Info("mutations enabled via grpc", "chat_target", chatGRPCTarget)
	return &MutationHandler{
		chatClient: chatv1.NewChatServiceClient(conn),
		chatConn:   conn,
	}
}

// Close closes the underlying gRPC connection.
func (h *MutationHandler) Close() {
	if h.chatConn != nil {
		h.chatConn.Close()
	}
}

// Handle dispatches a mutation message to the appropriate gRPC service.
// Runs in a goroutine to avoid blocking the hub event loop.
func (h *MutationHandler) Handle(client *Client, msg Message) {
	go func() {
		result, err := h.dispatch(client, msg)
		if err != nil {
			client.SendMessage(Message{
				Type:  Error,
				ID:    msg.ID,
				Error: err.Error(),
			})
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
	ctx := sdagrpc.ForwardJWT(context.Background(), client.JWT)

	switch msg.Action {
	case "create_session":
		var req chatv1.CreateSessionRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		req.UserId = client.UserID
		resp, err := h.chatClient.CreateSession(ctx, &req)
		if err != nil {
			return nil, err
		}
		return protojson.Marshal(resp)

	case "delete_session":
		var req chatv1.DeleteSessionRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		req.UserId = client.UserID
		resp, err := h.chatClient.DeleteSession(ctx, &req)
		if err != nil {
			return nil, err
		}
		return protojson.Marshal(resp)

	case "rename_session":
		var req chatv1.RenameSessionRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		req.UserId = client.UserID
		resp, err := h.chatClient.RenameSession(ctx, &req)
		if err != nil {
			return nil, err
		}
		return protojson.Marshal(resp)

	case "send_message":
		var req chatv1.AddMessageRequest
		if err := json.Unmarshal(msg.Data, &req); err != nil {
			return nil, err
		}
		req.UserId = client.UserID
		resp, err := h.chatClient.AddMessage(ctx, &req)
		if err != nil {
			return nil, err
		}
		return protojson.Marshal(resp)

	default:
		return nil, &unknownActionError{action: msg.Action}
	}
}

type unknownActionError struct{ action string }

func (e *unknownActionError) Error() string { return "unknown mutation action: " + e.action }
