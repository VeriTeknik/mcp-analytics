package config

import (
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config holds all configuration for the Analytics Service
type Config struct {
	// Service configuration
	Environment string `env:"ANALYTICS_ENV" envDefault:"development"`
	Port        int    `env:"ANALYTICS_PORT" envDefault:"8081"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`

	// Database URLs
	PostgresURL      string `env:"POSTGRES_URL" envDefault:"postgres://analytics:analytics@localhost:5432/mcp_analytics?sslmode=disable"`
	MongoDBURL       string `env:"MONGODB_URL" envDefault:"mongodb://localhost:27017"`
	MongoDBDatabase  string `env:"MONGODB_DATABASE" envDefault:"mcp_analytics"`
	ElasticsearchURL string `env:"ELASTICSEARCH_URL" envDefault:"http://localhost:9200"`
	RedisURL         string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`

	// Registry integration
	RegistryURL    string `env:"REGISTRY_URL" envDefault:"http://localhost:8080"`
	InternalAPIKey string `env:"INTERNAL_API_KEY" envDefault:"dev-internal-key"`

	// CORS configuration
	CORSOrigins string `env:"CORS_ORIGINS" envDefault:"http://localhost:3000,http://localhost:3001"`

	// Cache configuration
	CacheTTL                int `env:"CACHE_TTL" envDefault:"300"`        // 5 minutes
	SearchCacheTTL          int `env:"SEARCH_CACHE_TTL" envDefault:"300"` // 5 minutes
	FeaturedCacheTTL        int `env:"FEATURED_CACHE_TTL" envDefault:"900"` // 15 minutes
	TrendingCacheTTL        int `env:"TRENDING_CACHE_TTL" envDefault:"600"` // 10 minutes
	StatsCacheTTL           int `env:"STATS_CACHE_TTL" envDefault:"1800"` // 30 minutes

	// Rate limiting
	RateLimitEnabled        bool `env:"RATE_LIMIT_ENABLED" envDefault:"true"`
	RateLimitPublic         int  `env:"RATE_LIMIT_PUBLIC" envDefault:"100"`         // requests per minute
	RateLimitSearch         int  `env:"RATE_LIMIT_SEARCH" envDefault:"30"`          // requests per minute
	RateLimitAuthenticated  int  `env:"RATE_LIMIT_AUTHENTICATED" envDefault:"200"`  // requests per minute

	// Feature flags
	EnableRealTimeAnalytics bool `env:"ENABLE_REAL_TIME_ANALYTICS" envDefault:"true"`
	EnableSearchSuggestions bool `env:"ENABLE_SEARCH_SUGGESTIONS" envDefault:"true"`
	EnableWebSocket         bool `env:"ENABLE_WEBSOCKET" envDefault:"true"`

	// Monitoring
	PrometheusEnabled bool   `env:"PROMETHEUS_ENABLED" envDefault:"false"`
	OTELEndpoint      string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`

	// Security
	JWTSecret          string `env:"JWT_SECRET" envDefault:"dev-jwt-secret"`
	APIKeyHeader       string `env:"API_KEY_HEADER" envDefault:"X-API-Key"`
	InternalKeyHeader  string `env:"INTERNAL_KEY_HEADER" envDefault:"X-Internal-Key"`

	// Batch processing
	EventBatchSize     int `env:"EVENT_BATCH_SIZE" envDefault:"100"`
	EventFlushInterval int `env:"EVENT_FLUSH_INTERVAL" envDefault:"5"` // seconds

	// Search configuration
	SearchMaxResults    int `env:"SEARCH_MAX_RESULTS" envDefault:"100"`
	SearchDefaultLimit  int `env:"SEARCH_DEFAULT_LIMIT" envDefault:"20"`
	SearchMinQueryLen   int `env:"SEARCH_MIN_QUERY_LEN" envDefault:"2"`

	// Analytics configuration
	TrendingPeriodHours int     `env:"TRENDING_PERIOD_HOURS" envDefault:"168"` // 7 days
	TrendingMinInstalls int     `env:"TRENDING_MIN_INSTALLS" envDefault:"10"`
	MinRatingCount      int     `env:"MIN_RATING_COUNT" envDefault:"5"`
	PopularityDecayRate float64 `env:"POPULARITY_DECAY_RATE" envDefault:"0.95"`
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}
	
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate environment
	validEnvs := []string{"development", "staging", "production"}
	valid := false
	for _, env := range validEnvs {
		if c.Environment == env {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}

	// Validate port
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	// Validate URLs
	if c.PostgresURL == "" {
		return fmt.Errorf("PostgreSQL URL is required")
	}
	if c.MongoDBURL == "" {
		return fmt.Errorf("MongoDB URL is required")
	}
	if c.ElasticsearchURL == "" {
		return fmt.Errorf("Elasticsearch URL is required")
	}
	if c.RedisURL == "" {
		return fmt.Errorf("Redis URL is required")
	}
	if c.RegistryURL == "" {
		return fmt.Errorf("Registry URL is required")
	}

	// Validate API keys in production
	if c.Environment == "production" {
		if c.InternalAPIKey == "" || c.InternalAPIKey == "dev-internal-key" {
			return fmt.Errorf("internal API key must be set in production")
		}
		if c.JWTSecret == "" || c.JWTSecret == "dev-jwt-secret" {
			return fmt.Errorf("JWT secret must be set in production")
		}
	}

	return nil
}

// GetCORSOrigins returns CORS origins as a slice
func (c *Config) GetCORSOrigins() []string {
	if c.CORSOrigins == "" {
		return []string{}
	}
	
	origins := strings.Split(c.CORSOrigins, ",")
	for i, origin := range origins {
		origins[i] = strings.TrimSpace(origin)
	}
	
	return origins
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}