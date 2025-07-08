package model

import (
	"time"
)

// ServerDetail represents the complete server information
type ServerDetail struct {
	// Core fields from Registry
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Author       string              `json:"author,omitempty"`
	Homepage     string              `json:"homepage,omitempty"`
	Source       string              `json:"source"`
	Repository   string              `json:"repository,omitempty"`
	License      string              `json:"license,omitempty"`
	Categories   []string            `json:"categories,omitempty"`
	Packages     []Package           `json:"packages,omitempty"`
	VersionDetail VersionDetail      `json:"version_detail,omitempty"`
	Remotes      []Remote            `json:"remotes,omitempty"`
	
	// MCP capabilities
	Tools        []Capability        `json:"tools,omitempty"`
	Prompts      []Capability        `json:"prompts,omitempty"`
	Templates    []Capability        `json:"templates,omitempty"`
	
	// Analytics fields
	IndexedAt       time.Time          `json:"indexed_at"`
	LastUpdated     time.Time          `json:"last_updated"`
	InstallCount    int64              `json:"install_count"`
	RatingAverage   float64            `json:"rating_average"`
	RatingCount     int64              `json:"rating_count"`
	PopularityScore float64            `json:"popularity_score"`
	TrendingScore   float64            `json:"trending_score"`
	QualityScore    float64            `json:"quality_score"`
	
	// Search score (populated during search)
	Score           float64            `json:"score,omitempty"`
}

// Package represents a package distribution method
type Package struct {
	Type    string `json:"type"`    // npm, pypi, etc.
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// VersionDetail contains version information
type VersionDetail struct {
	Version         string `json:"version,omitempty"`
	SDKVersion      string `json:"sdk_version,omitempty"`
	ProtocolVersion string `json:"protocol_version,omitempty"`
}

// Remote represents a remote connection method
type Remote struct {
	Type      string            `json:"type"`      // stdio, http, sse
	Transport string            `json:"transport"` // stdio, http, sse
	Command   string            `json:"command,omitempty"`
	Args      []string          `json:"args,omitempty"`
	URL       string            `json:"url,omitempty"`
	Headers   map[string]string `json:"headers,omitempty"`
}

// Capability represents a tool, prompt, or template
type Capability struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// ServerStats represents aggregated statistics for a server
type ServerStats struct {
	ServerID         string    `json:"server_id"`
	InstallCount     int64     `json:"install_count"`
	RemoveCount      int64     `json:"remove_count"`
	RatingTotal      float64   `json:"rating_total"`
	RatingCount      int64     `json:"rating_count"`
	RatingAverage    float64   `json:"rating_average"`
	ToolCallCount    int64     `json:"tool_call_count"`
	PromptUseCount   int64     `json:"prompt_use_count"`
	TemplateUseCount int64     `json:"template_use_count"`
	LastCalculated   time.Time `json:"last_calculated"`
}