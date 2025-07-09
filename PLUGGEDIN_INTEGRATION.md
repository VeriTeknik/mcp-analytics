# plugged.in App - Analytics Integration Guide

This guide provides comprehensive documentation for integrating the plugged.in app with the MCP Analytics Service.

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [API Endpoints](#api-endpoints)
4. [Event Tracking](#event-tracking)
5. [Search Integration](#search-integration)
6. [Real-time Updates](#real-time-updates)
7. [SDK Usage](#sdk-usage)
8. [Best Practices](#best-practices)
9. [Error Handling](#error-handling)
10. [Examples](#examples)

## Overview

The Analytics Service provides comprehensive tracking, search, and discovery features for MCP servers in the plugged.in app.

### Base URLs
- **Production**: `https://analytics.plugged.in`

### API Versioning
All endpoints are versioned. Current version: `v1`

## Authentication

### Public Endpoints
Most read endpoints are public and don't require authentication:
- Search endpoints
- Discovery endpoints (featured, trending, etc.)
- Server analytics (read-only)
- Global statistics

### Authenticated Endpoints
User-specific actions require authentication:
- Install/uninstall tracking
- Rating submission
- Review posting
- Usage event tracking

### Authentication Methods

#### 1. API Key Authentication
For server-to-server communication:
```http
GET /v1/search
X-API-Key: your-api-key-here
```

#### 2. User Token Authentication
For user-specific actions:
```http
POST /v1/installs
Authorization: Bearer user-jwt-token
```

#### 3. Internal Authentication
For Registry-to-Analytics communication:
```http
POST /internal/events
X-Internal-Key: internal-shared-secret
```

## API Endpoints

### Search & Discovery

#### Search Servers
```http
GET /v1/search
```

Query Parameters:
- `q` (string): Search query
- `package_type` (array): Filter by package types (npm, pypi, docker, etc.)
- `transport_type` (array): Filter by transport (stdio, http, etc.)
- `source` (array): Filter by source (github, community, private)
- `category` (array): Filter by categories
- `min_rating` (number): Minimum rating (1-5)
- `has_tools` (boolean): Only servers with tools
- `has_prompts` (boolean): Only servers with prompts
- `has_resources` (boolean): Only servers with resources
- `sort` (string): Sort field (relevance, popularity, rating, installs, trending, recent)
- `order` (string): Sort order (asc, desc)
- `page` (number): Page number (1-based)
- `limit` (number): Results per page (max 100)

Example Request:
```bash
curl "https://analytics.plugged.in/v1/search?q=database&package_type=npm,docker&min_rating=4&sort=popularity&limit=20"
```

Example Response:
```json
{
  "servers": [
    {
      "id": "io.github.example/mcp-database",
      "name": "MCP Database Tools",
      "description": "Comprehensive database tools for MCP",
      "source": "github",
      "package_types": ["npm", "docker"],
      "transport_types": ["stdio"],
      "categories": ["database", "dev-tools"],
      "metrics": {
        "install_count": 15420,
        "active_installs": 8934,
        "monthly_active_users": 6234
      },
      "ratings": {
        "average": 4.7,
        "count": 234
      },
      "scores": {
        "popularity": 87.5,
        "trending": 12.3,
        "quality": 92.1,
        "relevance": 95.0
      },
      "capabilities": {
        "has_tools": true,
        "has_prompts": true,
        "has_resources": false,
        "tool_count": 12,
        "prompt_count": 5
      }
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total_results": 145,
    "total_pages": 8
  },
  "facets": {
    "package_types": {
      "npm": 89,
      "docker": 45,
      "pypi": 23
    },
    "categories": {
      "database": 34,
      "dev-tools": 56,
      "ai-tools": 23
    },
    "sources": {
      "github": 120,
      "community": 20,
      "private": 5
    }
  }
}
```

#### Featured Servers
```http
GET /v1/featured
```

Query Parameters:
- `category` (string): Filter by category
- `limit` (number): Number of results (default 10)

#### Trending Servers
```http
GET /v1/trending
```

Query Parameters:
- `period` (string): Time period (day, week, month)
- `category` (string): Filter by category
- `limit` (number): Number of results (default 20)

#### Top Rated Servers
```http
GET /v1/top-rated
```

Query Parameters:
- `category` (string): Filter by category
- `min_reviews` (number): Minimum number of reviews
- `limit` (number): Number of results (default 20)

#### Recently Active
```http
GET /v1/recent
```

Query Parameters:
- `activity_type` (string): Type of activity (updated, new)
- `limit` (number): Number of results (default 20)

### Server Analytics

#### Get Server Analytics
```http
GET /v1/servers/{id}/analytics
```

Response:
```json
{
  "server_id": "io.github.example/mcp-database",
  "metrics": {
    "total_installs": 15420,
    "active_installs": 8934,
    "daily_active_users": 1234,
    "weekly_active_users": 4567,
    "monthly_active_users": 6234,
    "uninstall_rate": 0.12,
    "retention_rate": {
      "day_1": 0.85,
      "day_7": 0.67,
      "day_30": 0.45
    }
  },
  "ratings": {
    "average": 4.7,
    "count": 234,
    "distribution": {
      "1": 5,
      "2": 8,
      "3": 21,
      "4": 67,
      "5": 133
    },
    "recent_trend": "improving"
  },
  "usage": {
    "tool_calls": {
      "total": 45678,
      "daily_average": 1523,
      "by_tool": {
        "query_database": 23456,
        "execute_sql": 12345,
        "backup_database": 9877
      },
      "success_rate": 0.94
    },
    "prompt_executions": {
      "total": 12345,
      "by_prompt": {
        "create_table": 5678,
        "optimize_query": 6667
      }
    }
  },
  "trending": {
    "rank": 5,
    "previous_rank": 8,
    "growth_rate": 0.23,
    "velocity": "rising"
  }
}
```

#### Get Server Reviews
```http
GET /v1/servers/{id}/reviews
```

Query Parameters:
- `sort` (string): Sort by (helpful, recent, rating)
- `verified_only` (boolean): Only verified installs
- `page` (number): Page number
- `limit` (number): Results per page

### User Interactions

#### Track Installation
```http
POST /v1/installs
Authorization: Bearer {user-token}
```

Request Body:
```json
{
  "server_id": "io.github.example/mcp-database",
  "platform": "macos",
  "app_version": "1.2.3",
  "package_type": "npm",
  "metadata": {
    "source": "search",
    "referrer": "trending_list"
  }
}
```

#### Track Uninstallation
```http
POST /v1/uninstalls
Authorization: Bearer {user-token}
```

Request Body:
```json
{
  "server_id": "io.github.example/mcp-database",
  "reason": "not_needed",
  "feedback": "Great tool but no longer needed for my project"
}
```

#### Submit Rating
```http
POST /v1/ratings
Authorization: Bearer {user-token}
```

Request Body:
```json
{
  "server_id": "io.github.example/mcp-database",
  "rating": 5,
  "review": {
    "title": "Excellent database tools!",
    "content": "This MCP server has transformed how I work with databases...",
    "pros": ["Fast", "Reliable", "Great documentation"],
    "cons": ["Learning curve"],
    "would_recommend": true
  }
}
```

#### Track Usage Event
```http
POST /v1/usage
Authorization: Bearer {user-token}
```

Request Body:
```json
{
  "server_id": "io.github.example/mcp-database",
  "events": [
    {
      "type": "tool_call",
      "name": "query_database",
      "timestamp": "2024-01-20T10:30:00Z",
      "duration_ms": 145,
      "success": true,
      "metadata": {
        "database_type": "postgresql",
        "query_complexity": "medium"
      }
    }
  ]
}
```

### Aggregated Statistics

#### Global Statistics
```http
GET /v1/stats/global
```

Response:
```json
{
  "total_servers": 1234,
  "total_installs": 567890,
  "active_users": 45678,
  "total_reviews": 12345,
  "by_category": {
    "dev-tools": 234,
    "ai-agents": 189,
    "databases": 145
  },
  "by_package_type": {
    "npm": 567,
    "docker": 345,
    "pypi": 234
  },
  "growth": {
    "servers_30d": 145,
    "installs_30d": 23456,
    "users_30d": 5678
  }
}
```

#### Category Statistics
```http
GET /v1/stats/categories
```

#### Popular Tools
```http
GET /v1/stats/tools
```

Response shows most-used tools across all servers.

## Event Tracking

### Client-Side Event Tracking

The plugged.in app should track these events:

#### Installation Events
```javascript
analytics.track('server_installed', {
  server_id: 'io.github.example/mcp-database',
  package_type: 'npm',
  source: 'search_results',
  search_query: 'database tools'
});
```

#### Search Events
```javascript
analytics.track('search_performed', {
  query: 'database tools',
  filters: {
    package_type: ['npm', 'docker'],
    min_rating: 4
  },
  results_count: 23
});
```

#### View Events
```javascript
analytics.track('server_viewed', {
  server_id: 'io.github.example/mcp-database',
  source: 'trending_list',
  position: 3
});
```

#### Usage Events
```javascript
analytics.track('tool_used', {
  server_id: 'io.github.example/mcp-database',
  tool_name: 'query_database',
  duration_ms: 145,
  success: true
});
```

## Search Integration

### Implementing Search UI

#### 1. Search with Autocomplete
```javascript
// Initialize search client
const searchClient = new AnalyticsClient({
  baseURL: 'https://analytics.plugged.in',
  apiKey: 'your-api-key'
});

// Autocomplete as user types
async function handleSearchInput(query) {
  if (query.length < 2) return;
  
  const suggestions = await searchClient.suggest({
    q: query,
    limit: 5
  });
  
  updateSearchSuggestions(suggestions);
}

// Full search
async function performSearch(query, filters) {
  const results = await searchClient.search({
    q: query,
    ...filters,
    limit: 20
  });
  
  displaySearchResults(results);
  updateFacets(results.facets);
}
```

#### 2. Faceted Filtering
```javascript
// Update search when filters change
function handleFilterChange(filterType, value) {
  const currentFilters = getActiveFilters();
  
  if (currentFilters[filterType].includes(value)) {
    // Remove filter
    currentFilters[filterType] = currentFilters[filterType]
      .filter(v => v !== value);
  } else {
    // Add filter
    currentFilters[filterType].push(value);
  }
  
  performSearch(currentQuery, currentFilters);
}
```

#### 3. Sorting
```javascript
// Handle sort change
function handleSortChange(sortField) {
  performSearch(currentQuery, currentFilters, {
    sort: sortField,
    order: sortField === 'name' ? 'asc' : 'desc'
  });
}
```

### Discovery Sections

#### Featured Servers Section
```javascript
async function loadFeaturedServers() {
  const featured = await searchClient.getFeatured({
    category: currentCategory,
    limit: 6
  });
  
  displayFeaturedSection(featured);
}
```

#### Trending Servers Section
```javascript
async function loadTrendingServers() {
  const trending = await searchClient.getTrending({
    period: 'week',
    limit: 10
  });
  
  displayTrendingSection(trending);
}
```

## Real-time Updates

### WebSocket Connection

```javascript
// Connect to real-time updates
const ws = new WebSocket('wss://analytics.plugged.in/v1/realtime');

ws.on('open', () => {
  // Subscribe to updates
  ws.send(JSON.stringify({
    action: 'subscribe',
    channels: ['trending', 'new_servers', 'ratings']
  }));
});

ws.on('message', (data) => {
  const update = JSON.parse(data);
  
  switch (update.type) {
    case 'trending_update':
      updateTrendingList(update.data);
      break;
    case 'new_server':
      showNewServerNotification(update.data);
      break;
    case 'rating_update':
      updateServerRating(update.data);
      break;
  }
});
```

### Server-Sent Events (Alternative)

```javascript
// For simpler one-way updates
const eventSource = new EventSource(
  'https://analytics.plugged.in/v1/events/stream'
);

eventSource.addEventListener('trending', (event) => {
  const data = JSON.parse(event.data);
  updateTrendingServers(data);
});
```

## SDK Usage

### JavaScript/TypeScript SDK

```typescript
import { AnalyticsClient } from '@pluggedin/analytics-sdk';

// Initialize client
const analytics = new AnalyticsClient({
  apiKey: process.env.ANALYTICS_API_KEY,
  environment: 'production'
});

// Search for servers
const searchResults = await analytics.search({
  query: 'database tools',
  filters: {
    packageType: ['npm'],
    minRating: 4
  },
  sort: 'popularity',
  limit: 20
});

// Track installation
await analytics.trackInstall({
  serverId: 'io.github.example/mcp-database',
  platform: getPlatform(),
  metadata: {
    source: 'search'
  }
});

// Submit rating
await analytics.submitRating({
  serverId: 'io.github.example/mcp-database',
  rating: 5,
  review: {
    title: 'Great tool!',
    content: 'Very useful for database operations'
  }
});

// Track usage
await analytics.trackUsage({
  serverId: 'io.github.example/mcp-database',
  events: [
    {
      type: 'tool_call',
      name: 'query_database',
      success: true,
      duration: 145
    }
  ]
});
```

### React Hooks

```typescript
import { useSearch, useServerAnalytics, useTrending } from '@pluggedin/analytics-react';

function SearchPage() {
  const { results, loading, error, search } = useSearch();
  
  const handleSearch = (query: string) => {
    search({
      q: query,
      filters: activeFilters
    });
  };
  
  return (
    <SearchInterface
      onSearch={handleSearch}
      results={results}
      loading={loading}
    />
  );
}

function ServerDetails({ serverId }: { serverId: string }) {
  const { analytics, loading } = useServerAnalytics(serverId);
  
  if (loading) return <Skeleton />;
  
  return (
    <ServerAnalytics
      metrics={analytics.metrics}
      ratings={analytics.ratings}
      usage={analytics.usage}
    />
  );
}
```

## Best Practices

### 1. Batch Event Tracking
Instead of sending individual events, batch them:

```javascript
const eventBatcher = new EventBatcher({
  batchSize: 50,
  flushInterval: 5000 // 5 seconds
});

// Events are automatically batched
eventBatcher.track('tool_used', { ... });
```

### 2. Error Handling
Always handle API errors gracefully:

```javascript
try {
  const results = await analytics.search({ ... });
} catch (error) {
  if (error.code === 'RATE_LIMITED') {
    showRateLimitMessage();
  } else if (error.code === 'NETWORK_ERROR') {
    showOfflineMessage();
    useCachedResults();
  } else {
    console.error('Search failed:', error);
    showGenericError();
  }
}
```

### 3. Caching
Implement client-side caching:

```javascript
const searchCache = new LRUCache({
  max: 100,
  ttl: 1000 * 60 * 5 // 5 minutes
});

async function cachedSearch(params) {
  const cacheKey = JSON.stringify(params);
  
  if (searchCache.has(cacheKey)) {
    return searchCache.get(cacheKey);
  }
  
  const results = await analytics.search(params);
  searchCache.set(cacheKey, results);
  
  return results;
}
```

### 4. Optimistic Updates
Update UI immediately for better UX:

```javascript
async function installServer(serverId) {
  // Update UI immediately
  updateUIAsInstalled(serverId);
  
  try {
    await analytics.trackInstall({ serverId });
  } catch (error) {
    // Revert UI on failure
    revertUIAsNotInstalled(serverId);
    showError('Installation tracking failed');
  }
}
```

### 5. Progressive Enhancement
Load data progressively:

```javascript
// Load critical data first
const [featured, trending] = await Promise.all([
  analytics.getFeatured({ limit: 6 }),
  analytics.getTrending({ limit: 10 })
]);

displayInitialContent(featured, trending);

// Load additional data in background
analytics.getTopRated({ limit: 20 }).then(displayTopRated);
analytics.getRecentlyUpdated({ limit: 20 }).then(displayRecent);
```

## Error Handling

### Error Response Format
All errors follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid search query",
    "details": {
      "field": "q",
      "reason": "Query too short"
    }
  }
}
```

### Common Error Codes
- `RATE_LIMITED` - Too many requests
- `INVALID_TOKEN` - Authentication failed
- `NOT_FOUND` - Resource not found
- `VALIDATION_ERROR` - Invalid request data
- `SERVER_ERROR` - Internal server error

### Retry Strategy
```javascript
async function retryableRequest(fn, maxRetries = 3) {
  let lastError;
  
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      lastError = error;
      
      // Don't retry client errors
      if (error.status >= 400 && error.status < 500) {
        throw error;
      }
      
      // Exponential backoff
      await sleep(Math.pow(2, i) * 1000);
    }
  }
  
  throw lastError;
}
```

## Examples

### Complete Search Implementation

```javascript
class MCPServerSearch {
  constructor() {
    this.analytics = new AnalyticsClient({ ... });
    this.currentFilters = {};
    this.currentSort = 'relevance';
  }
  
  async init() {
    // Load initial data
    await Promise.all([
      this.loadFeatured(),
      this.loadTrending(),
      this.loadCategories()
    ]);
  }
  
  async search(query) {
    try {
      const results = await this.analytics.search({
        q: query,
        ...this.currentFilters,
        sort: this.currentSort,
        limit: 20
      });
      
      this.displayResults(results);
      this.updateFacets(results.facets);
      
      // Track search
      this.analytics.track('search_performed', {
        query,
        filters: this.currentFilters,
        results_count: results.pagination.total_results
      });
    } catch (error) {
      this.handleSearchError(error);
    }
  }
  
  async installServer(serverId) {
    try {
      // Track installation
      await this.analytics.trackInstall({
        serverId,
        platform: this.getPlatform(),
        metadata: {
          source: 'search_results'
        }
      });
      
      // Update UI
      this.markAsInstalled(serverId);
      
      // Show success
      this.showNotification('Server installed successfully!');
    } catch (error) {
      this.showError('Failed to track installation');
    }
  }
  
  async rateServer(serverId, rating, review) {
    try {
      await this.analytics.submitRating({
        serverId,
        rating,
        review
      });
      
      // Update UI
      this.updateServerRating(serverId, rating);
      
      // Show success
      this.showNotification('Thank you for your feedback!');
    } catch (error) {
      this.showError('Failed to submit rating');
    }
  }
}
```

### Dashboard Implementation

```javascript
class AnalyticsDashboard {
  async loadDashboard() {
    const [global, trending, topRated, categories] = await Promise.all([
      this.analytics.getGlobalStats(),
      this.analytics.getTrending({ period: 'week', limit: 10 }),
      this.analytics.getTopRated({ limit: 10 }),
      this.analytics.getCategoryStats()
    ]);
    
    this.renderGlobalStats(global);
    this.renderTrendingChart(trending);
    this.renderTopRatedList(topRated);
    this.renderCategoryBreakdown(categories);
  }
  
  subscribeToRealtimeUpdates() {
    const ws = new WebSocket('wss://analytics.plugged.in/v1/realtime');
    
    ws.on('message', (event) => {
      const update = JSON.parse(event.data);
      
      switch (update.type) {
        case 'stats_update':
          this.updateGlobalStats(update.data);
          break;
        case 'trending_change':
          this.updateTrendingChart(update.data);
          break;
      }
    });
  }
}
```

## Rate Limiting

The API implements rate limiting:
- **Public endpoints**: 100 requests/minute
- **Search endpoint**: 30 requests/minute
- **Authenticated endpoints**: 200 requests/minute

Rate limit headers:
```http
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642694400
```

## Support

For technical support or questions:
- Email: analytics-support@plugged.in
- Documentation: https://docs.plugged.in/analytics
- API Status: https://status.plugged.in

---

Last Updated: January 2024
API Version: v1