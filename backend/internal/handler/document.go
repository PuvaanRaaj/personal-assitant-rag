package handler

import (
	"github.com/PuvaanRaaj/personal-rag-agent/internal/middleware"
	"github.com/PuvaanRaaj/personal-rag-agent/internal/service"
	"github.com/gofiber/fiber/v2"
)

// DocumentHandler handles document requests
type DocumentHandler struct {
	documentService *service.DocumentService
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(documentService *service.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

// Upload handles document upload
func (h *DocumentHandler) Upload(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// TODO: Implement file upload logic
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "upload not implemented yet",
	})
}

// List handles listing user documents
func (h *DocumentHandler) List(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	// TODO: Implement list documents
	return c.JSON(fiber.Map{
		"documents": []interface{}{},
	})
}

// Get handles getting a single document
func (h *DocumentHandler) Get(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	documentID := c.Params("id")
	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "document ID is required",
		})
	}

	// TODO: Implement get document
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "get document not implemented yet",
	})
}

// Delete handles deleting a document
func (h *DocumentHandler) Delete(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "unauthorized",
		})
	}

	documentID := c.Params("id")
	if documentID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "document ID is required",
		})
	}

	// TODO: Implement delete document
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "delete document not implemented yet",
	})
}
