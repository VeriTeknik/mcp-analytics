package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pluggedin/mcp-analytics/internal/model"
	"github.com/pluggedin/mcp-analytics/internal/search"
)

// EventType represents the type of notification event
type EventType string

const (
	EventServerAdded   EventType = "server_added"
	EventServerUpdated EventType = "server_updated"
	EventServerDeleted EventType = "server_deleted"
)

// Event represents a notification event from Registry
type Event struct {
	Type      EventType              `json:"type"`
	ServerID  string                 `json:"server_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// EventHandler handles internal event notifications from Registry
type EventHandler struct {
	searchService *search.Service
	eventQueue    chan Event
}

// NewEventHandler creates a new event handler
func NewEventHandler(searchService *search.Service) *EventHandler {
	h := &EventHandler{
		searchService: searchService,
		eventQueue:    make(chan Event, 1000), // Buffer up to 1000 events
	}

	// Start event processor
	go h.processEvents()

	return h
}

// HandleEvent processes incoming events from Registry
func (h *EventHandler) HandleEvent(c *fiber.Ctx) error {
	var event Event
	if err := c.BodyParser(&event); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid event format",
		})
	}

	// Validate event
	if event.Type == "" || event.ServerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	// Queue event for processing
	select {
	case h.eventQueue <- event:
		log.Printf("Event queued: %s for server %s", event.Type, event.ServerID)
	default:
		log.Printf("Event queue full, dropping event: %s for server %s", event.Type, event.ServerID)
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Event queue full",
		})
	}

	return c.JSON(fiber.Map{
		"status": "accepted",
	})
}

// processEvents processes events from the queue
func (h *EventHandler) processEvents() {
	for event := range h.eventQueue {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		switch event.Type {
		case EventServerAdded:
			h.handleServerAdded(ctx, event)
		case EventServerUpdated:
			h.handleServerUpdated(ctx, event)
		case EventServerDeleted:
			h.handleServerDeleted(ctx, event)
		default:
			log.Printf("Unknown event type: %s", event.Type)
		}
		
		cancel()
	}
}

// handleServerAdded processes server added events
func (h *EventHandler) handleServerAdded(ctx context.Context, event Event) {
	log.Printf("Processing server added: %s", event.ServerID)

	// Convert event data to ServerDetail
	serverDetail, err := h.eventDataToServerDetail(event.Data)
	if err != nil {
		log.Printf("Failed to parse server data: %v", err)
		return
	}

	// Index in Elasticsearch
	if err := h.searchService.IndexServer(ctx, serverDetail); err != nil {
		log.Printf("Failed to index server %s: %v", event.ServerID, err)
		return
	}

	log.Printf("Successfully indexed server: %s", event.ServerID)
}

// handleServerUpdated processes server updated events
func (h *EventHandler) handleServerUpdated(ctx context.Context, event Event) {
	log.Printf("Processing server updated: %s", event.ServerID)

	// For updates, we might get partial data, so fetch the full server first
	server, err := h.searchService.GetServer(ctx, event.ServerID)
	if err != nil {
		log.Printf("Failed to get existing server %s: %v", event.ServerID, err)
		return
	}

	// Apply updates from event data
	if err := h.applyUpdates(server, event.Data); err != nil {
		log.Printf("Failed to apply updates: %v", err)
		return
	}

	// Re-index in Elasticsearch
	if err := h.searchService.IndexServer(ctx, server); err != nil {
		log.Printf("Failed to update server %s: %v", event.ServerID, err)
		return
	}

	log.Printf("Successfully updated server: %s", event.ServerID)
}

// handleServerDeleted processes server deleted events
func (h *EventHandler) handleServerDeleted(ctx context.Context, event Event) {
	log.Printf("Processing server deleted: %s", event.ServerID)

	// Delete from Elasticsearch
	if err := h.searchService.DeleteServer(ctx, event.ServerID); err != nil {
		log.Printf("Failed to delete server %s: %v", event.ServerID, err)
		return
	}

	log.Printf("Successfully deleted server: %s", event.ServerID)
}

// eventDataToServerDetail converts event data to ServerDetail model
func (h *EventHandler) eventDataToServerDetail(data map[string]interface{}) (*model.ServerDetail, error) {
	// Marshal to JSON then unmarshal to ServerDetail for proper type conversion
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal event data: %w", err)
	}

	var server model.ServerDetail
	if err := json.Unmarshal(jsonData, &server); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to ServerDetail: %w", err)
	}

	// Set additional fields
	server.IndexedAt = time.Now()
	server.LastUpdated = time.Now()

	return &server, nil
}

// applyUpdates applies updates from event data to existing server
func (h *EventHandler) applyUpdates(server *model.ServerDetail, updates map[string]interface{}) error {
	// For now, we'll just re-parse the entire object
	// In the future, we might want to do field-by-field updates
	updatedServer, err := h.eventDataToServerDetail(updates)
	if err != nil {
		return err
	}

	// Preserve certain fields from the original
	updatedServer.ID = server.ID
	updatedServer.IndexedAt = server.IndexedAt
	updatedServer.LastUpdated = time.Now()

	// Copy updated fields back
	*server = *updatedServer

	return nil
}

// InternalAuthMiddleware validates internal API key
func InternalAuthMiddleware(apiKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := c.Get("X-Internal-Key")
		if key == "" || key != apiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized",
			})
		}
		return c.Next()
	}
}