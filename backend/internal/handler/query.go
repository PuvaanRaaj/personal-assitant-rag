package handler

import (
	"github.com/PuvaanRaaj/personal-rag-agent/internal/middleware"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/service"
	"github.com/gofiber/fiber/v2"
)

// QueryHandler handles query requests
type QueryHandler struct {
	ragService *service.RAGService
}

// NewQueryHandler creates a new query handler
func NewQueryHandler(ragService *service.RAGService) *QueryHandler {
	return &QueryHandler{ragService: ragService}
}

// QueryRequest represents a query request
type QueryRequest struct {
	Question string `json:"question" validate:"required"`
}

// Query handles RAG queries
func (h *QueryHandler) Query(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	var req QueryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if req.Question == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "question is required",
		})
	}

	// Perform RAG query
	response, err := h.ragService.Query(c.Context(), userID, req.Question)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(response)
}

// StreamQuery handles streaming RAG queries
func (h *QueryHandler) StreamQuery(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// TODO: Implement streaming query with SSE
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "streaming query not implemented yet",
	})
}
