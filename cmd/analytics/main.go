package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/pluggedin/mcp-analytics/internal/api"
	"github.com/pluggedin/mcp-analytics/internal/config"
	"github.com/pluggedin/mcp-analytics/internal/search"
)

var (
	// Version information (set during build)
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	log.Printf("Starting MCP Analytics Service v%s (build: %s)", Version, BuildTime)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize search service
	log.Println("Initializing search service...")
	searchService, err := search.NewService(cfg.ElasticsearchURL)
	if err != nil {
		log.Fatalf("Failed to initialize search service: %v", err)
	}

	// Create event handler
	eventHandler := api.NewEventHandler(searchService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:           "MCP Analytics Service",
		EnablePrintRoutes: cfg.Environment == "development",
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(cfg.GetCORSOrigins(), ","),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-API-Key,X-Internal-Key",
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"version": Version,
			"time":    time.Now().UTC(),
		})
	})

	// Ready check endpoint
	app.Get("/ready", func(c *fiber.Ctx) error {
		// TODO: Add dependency checks
		return c.JSON(fiber.Map{
			"status": "ready",
		})
	})

	// Internal API routes (protected by internal key)
	internal := app.Group("/internal", api.InternalAuthMiddleware(cfg.InternalAPIKey))
	internal.Post("/events", eventHandler.HandleEvent)

	// Public API routes
	v1 := app.Group("/v1")

	// Search endpoint
	v1.Get("/search", func(c *fiber.Ctx) error {
		query := search.SearchQuery{
			Query:   c.Query("q"),
			Sort:    c.Query("sort", "relevance"),
			Offset:  c.QueryInt("offset", 0),
			Limit:   c.QueryInt("limit", 20),
			Filters: make(map[string]interface{}),
		}

		// Add filters
		if pkgType := c.Query("package_type"); pkgType != "" {
			query.Filters["package_type"] = pkgType
		}
		if transport := c.Query("transport"); transport != "" {
			query.Filters["transport"] = transport
		}
		if category := c.Query("category"); category != "" {
			query.Filters["categories"] = category
		}
		if source := c.Query("source"); source != "" {
			query.Filters["source"] = source
		}

		// Validate limit
		if query.Limit > cfg.SearchMaxResults {
			query.Limit = cfg.SearchMaxResults
		}

		// Execute search
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()

		result, err := searchService.Search(ctx, query)
		if err != nil {
			log.Printf("Search error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Search failed",
			})
		}

		return c.JSON(result)
	})

	// Start server in goroutine
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Port)
		log.Printf("Starting HTTP server on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}