package handler

import (
	"context"
	"log/slog"

	chatv1 "github.com/Camionerou/rag-saldivia/gen/go/chat/v1"
	sdamw "github.com/Camionerou/rag-saldivia/pkg/middleware"
	"github.com/Camionerou/rag-saldivia/services/chat/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GRPCHandler implements the ChatService gRPC server.
type GRPCHandler struct {
	chatv1.UnimplementedChatServiceServer
	svc *service.Chat
}

// NewGRPC creates a chat gRPC handler backed by the same service layer as HTTP.
func NewGRPC(svc *service.Chat) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

func (h *GRPCHandler) CreateSession(ctx context.Context, req *chatv1.CreateSessionRequest) (*chatv1.Session, error) {
	userID := sdamw.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	var collection *string
	if req.Collection != nil {
		collection = req.Collection
	}

	s, err := h.svc.CreateSession(ctx, userID, req.Title, collection)
	if err != nil {
		slog.Error("grpc create session failed", "error", err)
		return nil, status.Error(codes.Internal, "create session failed")
	}
	return sessionToProto(s), nil
}

func (h *GRPCHandler) GetSession(ctx context.Context, req *chatv1.GetSessionRequest) (*chatv1.Session, error) {
	userID := sdamw.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	s, err := h.svc.GetSession(ctx, req.SessionId, userID)
	if err != nil {
		if err == service.ErrSessionNotFound {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		return nil, status.Error(codes.Internal, "get session failed")
	}
	return sessionToProto(s), nil
}

func (h *GRPCHandler) ListSessions(ctx context.Context, req *chatv1.ListSessionsRequest) (*chatv1.ListSessionsResponse, error) {
	userID := sdamw.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	sessions, err := h.svc.ListSessions(ctx, userID, 50, 0)
	if err != nil {
		return nil, status.Error(codes.Internal, "list sessions failed")
	}

	resp := &chatv1.ListSessionsResponse{}
	for _, s := range sessions {
		resp.Sessions = append(resp.Sessions, sessionToProto(&s))
	}
	return resp, nil
}

func (h *GRPCHandler) DeleteSession(ctx context.Context, req *chatv1.DeleteSessionRequest) (*chatv1.DeleteSessionResponse, error) {
	userID := sdamw.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	if err := h.svc.DeleteSession(ctx, req.SessionId, userID); err != nil {
		if err == service.ErrSessionNotFound {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		return nil, status.Error(codes.Internal, "delete session failed")
	}
	return &chatv1.DeleteSessionResponse{}, nil
}

func (h *GRPCHandler) RenameSession(ctx context.Context, req *chatv1.RenameSessionRequest) (*chatv1.RenameSessionResponse, error) {
	userID := sdamw.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	if err := h.svc.RenameSession(ctx, req.SessionId, userID, req.Title); err != nil {
		if err == service.ErrSessionNotFound {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		return nil, status.Error(codes.Internal, "rename session failed")
	}
	return &chatv1.RenameSessionResponse{}, nil
}

func (h *GRPCHandler) AddMessage(ctx context.Context, req *chatv1.AddMessageRequest) (*chatv1.Message, error) {
	userID := sdamw.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	// Validate role (parity with HTTP handler)
	switch req.Role {
	case "user", "assistant":
		// ok
	case "system":
		return nil, status.Error(codes.PermissionDenied, "system role not allowed from API")
	default:
		return nil, status.Error(codes.InvalidArgument, "invalid role")
	}

	m, err := h.svc.AddMessage(ctx, req.SessionId, userID, req.Role, req.Content, nil, req.Sources, req.Metadata)
	if err != nil {
		slog.Error("grpc add message failed", "error", err)
		return nil, status.Error(codes.Internal, "add message failed")
	}
	return messageToProto(m), nil
}

func (h *GRPCHandler) ListMessages(ctx context.Context, req *chatv1.ListMessagesRequest) (*chatv1.ListMessagesResponse, error) {
	userID := sdamw.UserIDFromContext(ctx)
	if userID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing user context")
	}

	// Verify session ownership
	if _, err := h.svc.GetSession(ctx, req.SessionId, userID); err != nil {
		if err == service.ErrSessionNotFound {
			return nil, status.Error(codes.NotFound, "session not found")
		}
		return nil, status.Error(codes.Internal, "get session failed")
	}

	messages, err := h.svc.GetMessages(ctx, req.SessionId, 100)
	if err != nil {
		return nil, status.Error(codes.Internal, "list messages failed")
	}

	resp := &chatv1.ListMessagesResponse{}
	for _, m := range messages {
		resp.Messages = append(resp.Messages, messageToProto(&m))
	}
	return resp, nil
}

func sessionToProto(s *service.Session) *chatv1.Session {
	ps := &chatv1.Session{
		Id:        s.ID,
		UserId:    s.UserID,
		Title:     s.Title,
		IsSaved:   s.IsSaved,
		CreatedAt: timestamppb.New(s.CreatedAt),
		UpdatedAt: timestamppb.New(s.UpdatedAt),
	}
	if s.Collection != nil {
		ps.Collection = s.Collection
	}
	return ps
}

func messageToProto(m *service.Message) *chatv1.Message {
	return &chatv1.Message{
		Id:        m.ID,
		SessionId: m.SessionID,
		Role:      m.Role,
		Content:   m.Content,
		Sources:   m.Sources,
		Metadata:  m.Metadata,
		CreatedAt: timestamppb.New(m.CreatedAt),
	}
}
