package handler

import (
	"context"
	"log/slog"

	searchv1 "github.com/Camionerou/rag-saldivia/gen/go/search/v1"
	"github.com/Camionerou/rag-saldivia/services/search/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCHandler implements the SearchService gRPC server.
type GRPCHandler struct {
	searchv1.UnimplementedSearchServiceServer
	svc *service.Search
}

// NewGRPC creates a search gRPC handler backed by the same service layer as HTTP.
func NewGRPC(svc *service.Search) *GRPCHandler {
	return &GRPCHandler{svc: svc}
}

// Query performs a 3-phase tree search across document collections.
func (h *GRPCHandler) Query(ctx context.Context, req *searchv1.SearchRequest) (*searchv1.SearchResponse, error) {
	if req.Query == "" {
		return nil, status.Error(codes.InvalidArgument, "query is required")
	}

	collectionID := ""
	if req.CollectionId != nil {
		collectionID = *req.CollectionId
	}

	result, err := h.svc.SearchDocuments(ctx, req.Query, collectionID, int(req.MaxNodes))
	if err != nil {
		slog.Error("grpc search failed", "error", err)
		return nil, status.Error(codes.Internal, "search failed")
	}

	return toProto(result), nil
}

func toProto(r *service.SearchResult) *searchv1.SearchResponse {
	resp := &searchv1.SearchResponse{
		Query:      r.Query,
		DurationMs: int32(r.DurationMS),
	}
	for _, s := range r.Selections {
		pages := make([]int32, len(s.Pages))
		for i, p := range s.Pages {
			pages[i] = int32(p)
		}
		resp.Selections = append(resp.Selections, &searchv1.Selection{
			Document:   s.Document,
			DocumentId: s.DocumentID,
			NodeIds:    s.NodeIDs,
			Sections:   s.Sections,
			Pages:      pages,
			Text:       s.Text,
			Tables:     s.Tables,
			Images:     s.Images,
		})
	}
	return resp
}
