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

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "no file uploaded",
		})
	}

	// Process document
	doc, err := h.documentService.UploadDocument(c.Context(), userID, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "document uploaded successfully",
		"document": doc,
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

	documents, err := h.documentService.ListDocuments(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list documents",
		})
	}

	return c.JSON(fiber.Map{
		"documents": documents,
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

	doc, err := h.documentService.GetDocument(c.Context(), userID, documentID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"document": doc,
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

	if err := h.documentService.DeleteDocument(c.Context(), userID, documentID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "document deleted successfully",
	})
}
