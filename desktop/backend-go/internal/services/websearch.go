package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
)

// ============================================================================
// TYPES AND INTERFACES
// ============================================================================

// WebSearchResult represents a single search result
type WebSearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	Source      string `json:"source"`
	PublishedAt string `json:"published_at,omitempty"`
	Score       float64 `json:"score,omitempty"` // Relevance score if available
}

// WebSearchResponse represents the full search response
type WebSearchResponse struct {
	Query        string            `json:"query"`
	Results      []WebSearchResult `json:"results"`
	TotalResults int               `json:"total_results"`
	SearchTime   float64           `json:"search_time_ms"`
	Provider     string            `json:"provider"`
	Cached       bool              `json:"cached,omitempty"`
}

// SearchProvider interface for different search backends
type SearchProvider interface {
	Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error)
	Name() string
	Available() bool
}

// ============================================================================
// BRAVE SEARCH PROVIDER
// ============================================================================

// BraveSearchProvider implements search using Brave Search API
// Free tier: 2000 queries/month
// Docs: https://brave.com/search/api/
type BraveSearchProvider struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

// NewBraveSearchProvider creates a new Brave Search provider
func NewBraveSearchProvider(apiKey string) *BraveSearchProvider {
	return &BraveSearchProvider{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: "https://api.search.brave.com/res/v1/web/search",
	}
}

func (b *BraveSearchProvider) Name() string { return "brave" }

func (b *BraveSearchProvider) Available() bool { return b.apiKey != "" }

func (b *BraveSearchProvider) Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error) {
	if !b.Available() {
		return nil, fmt.Errorf("brave search API key not configured")
	}

	startTime := time.Now()

	// Build request URL
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", min(maxResults, 20)))
	params.Set("text_decorations", "false")
	params.Set("search_lang", "en")
	params.Set("safesearch", "moderate")

	reqURL := b.baseURL + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("X-Subscription-Token", b.apiKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("brave API returned status %d: %s", resp.StatusCode, string(body))
	}

	var braveResp struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
				PageAge     string `json:"page_age,omitempty"`
			} `json:"results"`
		} `json:"web"`
		Query struct {
			Original string `json:"original"`
		} `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&braveResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make([]WebSearchResult, 0, len(braveResp.Web.Results))
	for _, r := range braveResp.Web.Results {
		results = append(results, WebSearchResult{
			Title:       r.Title,
			URL:         r.URL,
			Snippet:     r.Description,
			Source:      extractDomain(r.URL),
			PublishedAt: r.PageAge,
		})
	}

	return &WebSearchResponse{
		Query:        query,
		Results:      results,
		TotalResults: len(results),
		SearchTime:   float64(time.Since(startTime).Milliseconds()),
		Provider:     "brave",
	}, nil
}

// ============================================================================
// SERPER PROVIDER (Google Results)
// ============================================================================

// SerperProvider implements search using Serper.dev API (Google results)
// Free tier: 2500 queries/month
// Docs: https://serper.dev/
type SerperProvider struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

// NewSerperProvider creates a new Serper provider
func NewSerperProvider(apiKey string) *SerperProvider {
	return &SerperProvider{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: "https://google.serper.dev/search",
	}
}

func (s *SerperProvider) Name() string { return "serper" }

func (s *SerperProvider) Available() bool { return s.apiKey != "" }

func (s *SerperProvider) Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error) {
	if !s.Available() {
		return nil, fmt.Errorf("serper API key not configured")
	}

	startTime := time.Now()

	// Build request body
	reqBody := map[string]interface{}{
		"q":   query,
		"num": min(maxResults, 20),
		"gl":  "us",
		"hl":  "en",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-KEY", s.apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("serper API returned status %d: %s", resp.StatusCode, string(body))
	}

	var serperResp struct {
		Organic []struct {
			Title    string `json:"title"`
			Link     string `json:"link"`
			Snippet  string `json:"snippet"`
			Position int    `json:"position"`
			Date     string `json:"date,omitempty"`
		} `json:"organic"`
		SearchParameters struct {
			Q string `json:"q"`
		} `json:"searchParameters"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&serperResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make([]WebSearchResult, 0, len(serperResp.Organic))
	for _, r := range serperResp.Organic {
		results = append(results, WebSearchResult{
			Title:       r.Title,
			URL:         r.Link,
			Snippet:     r.Snippet,
			Source:      extractDomain(r.Link),
			PublishedAt: r.Date,
		})
	}

	return &WebSearchResponse{
		Query:        query,
		Results:      results,
		TotalResults: len(results),
		SearchTime:   float64(time.Since(startTime).Milliseconds()),
		Provider:     "serper",
	}, nil
}

// ============================================================================
// TAVILY PROVIDER (AI-Focused Search)
// ============================================================================

// TavilyProvider implements search using Tavily API (AI-optimized)
// Free tier: 1000 queries/month
// Docs: https://tavily.com/
type TavilyProvider struct {
	apiKey  string
	client  *http.Client
	baseURL string
}

// NewTavilyProvider creates a new Tavily provider
func NewTavilyProvider(apiKey string) *TavilyProvider {
	return &TavilyProvider{
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 30 * time.Second}, // Tavily can be slow
		baseURL: "https://api.tavily.com/search",
	}
}

func (t *TavilyProvider) Name() string { return "tavily" }

func (t *TavilyProvider) Available() bool { return t.apiKey != "" }

func (t *TavilyProvider) Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error) {
	if !t.Available() {
		return nil, fmt.Errorf("tavily API key not configured")
	}

	startTime := time.Now()

	// Build request body
	reqBody := map[string]interface{}{
		"api_key":             t.apiKey,
		"query":               query,
		"max_results":         min(maxResults, 10), // Tavily max is 10
		"search_depth":        "basic",            // "basic" or "advanced"
		"include_answer":      false,
		"include_raw_content": false,
		"include_images":      false,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tavily API returned status %d: %s", resp.StatusCode, string(body))
	}

	var tavilyResp struct {
		Results []struct {
			Title   string  `json:"title"`
			URL     string  `json:"url"`
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"results"`
		Query string `json:"query"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tavilyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	results := make([]WebSearchResult, 0, len(tavilyResp.Results))
	for _, r := range tavilyResp.Results {
		results = append(results, WebSearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Content,
			Source:  extractDomain(r.URL),
			Score:   r.Score,
		})
	}

	return &WebSearchResponse{
		Query:        query,
		Results:      results,
		TotalResults: len(results),
		SearchTime:   float64(time.Since(startTime).Milliseconds()),
		Provider:     "tavily",
	}, nil
}

// ============================================================================
// DUCKDUCKGO PROVIDER (Fallback - No API Key Required)
// ============================================================================

// userAgents is a list of common browser user agents for rotation
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0",
}

// DuckDuckGoProvider implements search using DuckDuckGo (no API key required)
type DuckDuckGoProvider struct {
	client *http.Client
}

// NewDuckDuckGoProvider creates a new DuckDuckGo provider
func NewDuckDuckGoProvider() *DuckDuckGoProvider {
	return &DuckDuckGoProvider{
		client: &http.Client{
			Timeout: 15 * time.Second,
			// Don't follow redirects automatically
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (d *DuckDuckGoProvider) Name() string { return "duckduckgo" }

func (d *DuckDuckGoProvider) Available() bool { return true } // Always available

func (d *DuckDuckGoProvider) Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error) {
	startTime := time.Now()

	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 20 {
		maxResults = 20
	}

	slog.Debug("DuckDuckGo search starting", "query", query, "maxResults", maxResults)

	// Try the Lite HTML version first (more reliable)
	results, err := d.searchLite(ctx, query, maxResults)
	if err != nil || len(results) == 0 {
		if err != nil {
			slog.Debug("DuckDuckGo Lite failed with error, trying HTML", "err", err)
		} else {
			slog.Debug("DuckDuckGo Lite returned 0 results, trying HTML")
		}
		// Fallback to standard HTML
		results, err = d.searchHTML(ctx, query, maxResults)
		if err != nil || len(results) == 0 {
			if err != nil {
				slog.Debug("DuckDuckGo HTML failed with error, trying instant", "err", err)
			} else {
				slog.Debug("DuckDuckGo HTML returned 0 results, trying instant")
			}
			// Final fallback to instant answers
			results, err = d.searchInstant(ctx, query)
			if err != nil {
				slog.Error("All DuckDuckGo search methods failed", "err", err, "query", query)
				return nil, fmt.Errorf("all DuckDuckGo search methods failed: %w", err)
			}
		}
	}

	slog.Debug("DuckDuckGo search completed", "query", query, "results", len(results), "duration", time.Since(startTime))

	return &WebSearchResponse{
		Query:        query,
		Results:      results,
		TotalResults: len(results),
		SearchTime:   float64(time.Since(startTime).Milliseconds()),
		Provider:     "duckduckgo",
	}, nil
}

// searchLite uses the Lite (no JS) version of DuckDuckGo
func (d *DuckDuckGoProvider) searchLite(ctx context.Context, query string, maxResults int) ([]WebSearchResult, error) {
	searchURL := "https://lite.duckduckgo.com/lite/"

	// POST request with form data
	formData := url.Values{}
	formData.Set("q", query)
	formData.Set("kl", "us-en") // US English

	req, err := http.NewRequestWithContext(ctx, "POST", searchURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return d.parseLiteHTML(string(body), maxResults), nil
}

// parseLiteHTML parses the Lite DuckDuckGo results page
func (d *DuckDuckGoProvider) parseLiteHTML(html string, maxResults int) []WebSearchResult {
	var results []WebSearchResult

	// DuckDuckGo Lite has a table-based structure
	// Results are in table rows with links followed by snippets

	// More flexible pattern - look for links that have uddg= (DuckDuckGo redirect URLs)
	// or result-link class, attributes can be in any order
	linkPatterns := []*regexp.Regexp{
		// Pattern 1: Links with uddg parameter (redirect URLs)
		regexp.MustCompile(`<a[^>]*href="([^"]*uddg=[^"]+)"[^>]*>([^<]+)</a>`),
		// Pattern 2: Links with result-link class
		regexp.MustCompile(`<a[^>]*class="[^"]*result-link[^"]*"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`),
		// Pattern 3: class before href
		regexp.MustCompile(`<a[^>]*href="([^"]+)"[^>]*class="[^"]*result-link[^"]*"[^>]*>([^<]+)</a>`),
		// Pattern 4: rel="nofollow" links (common for external links)
		regexp.MustCompile(`<a[^>]*rel="nofollow"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`),
	}

	// Try to find snippets - they appear in various formats
	snippetPatterns := []*regexp.Regexp{
		regexp.MustCompile(`<td[^>]*class="[^"]*result-snippet[^"]*"[^>]*>([^<]+)</td>`),
		regexp.MustCompile(`<span[^>]*class="[^"]*result-snippet[^"]*"[^>]*>([^<]+)</span>`),
		regexp.MustCompile(`class="[^"]*snippet[^"]*"[^>]*>([^<]+)<`),
	}

	// Collect all link matches from all patterns
	var allLinks [][]string
	seenURLs := make(map[string]bool)

	for _, pattern := range linkPatterns {
		matches := pattern.FindAllStringSubmatch(html, maxResults*3)
		for _, match := range matches {
			if len(match) >= 3 {
				rawURL := match[1]
				// Deduplicate
				if !seenURLs[rawURL] {
					seenURLs[rawURL] = true
					allLinks = append(allLinks, match)
				}
			}
		}
	}

	// Collect snippets
	var snippets []string
	for _, pattern := range snippetPatterns {
		matches := pattern.FindAllStringSubmatch(html, maxResults*3)
		for _, match := range matches {
			if len(match) >= 2 {
				snippets = append(snippets, match[1])
			}
		}
	}

	slog.Debug("DuckDuckGo Lite parsing", "linksFound", len(allLinks), "snippetsFound", len(snippets))

	snippetIdx := 0
	for _, match := range allLinks {
		if len(results) >= maxResults {
			break
		}

		rawURL := match[1]
		title := strings.TrimSpace(match[2])

		// Skip DuckDuckGo's internal navigation links
		if strings.Contains(rawURL, "duckduckgo.com") && !strings.Contains(rawURL, "uddg=") {
			continue
		}

		// Skip empty titles
		if title == "" || title == "..." {
			continue
		}

		// Clean the URL (handle DuckDuckGo redirect URLs)
		actualURL := cleanDDGURL(rawURL)
		if actualURL == "" || !strings.HasPrefix(actualURL, "http") {
			continue
		}

		// Skip DuckDuckGo pages
		if strings.Contains(actualURL, "duckduckgo.com") {
			continue
		}

		// Get snippet if available
		snippet := ""
		if snippetIdx < len(snippets) {
			snippet = cleanHTMLText(snippets[snippetIdx])
			snippetIdx++
		}

		results = append(results, WebSearchResult{
			Title:   cleanHTMLText(title),
			URL:     actualURL,
			Snippet: snippet,
			Source:  extractDomain(actualURL),
		})
	}

	slog.Debug("DuckDuckGo Lite results", "count", len(results))

	return results
}

// searchHTML uses the standard HTML version
func (d *DuckDuckGoProvider) searchHTML(ctx context.Context, query string, maxResults int) ([]WebSearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("DNT", "1")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return d.parseHTML(string(body), maxResults), nil
}

// parseHTML parses the HTML response from DuckDuckGo
func (d *DuckDuckGoProvider) parseHTML(html string, maxResults int) []WebSearchResult {
	var results []WebSearchResult

	// Multiple patterns to handle different HTML structures
	linkPatterns := []*regexp.Regexp{
		// Pattern for result__a class (standard html.duckduckgo.com)
		regexp.MustCompile(`<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`),
		regexp.MustCompile(`<a[^>]*href="([^"]+)"[^>]*class="[^"]*result__a[^"]*"[^>]*>([^<]+)</a>`),
		// Pattern for uddg redirect URLs
		regexp.MustCompile(`<a[^>]*href="([^"]*uddg=[^"]+)"[^>]*>([^<]+)</a>`),
		// Pattern for result-link class
		regexp.MustCompile(`<a[^>]*class="[^"]*result-link[^"]*"[^>]*href="([^"]+)"[^>]*>([^<]+)</a>`),
	}

	snippetPatterns := []*regexp.Regexp{
		regexp.MustCompile(`<a[^>]*class="[^"]*result__snippet[^"]*"[^>]*>([^<]+)</a>`),
		regexp.MustCompile(`class="[^"]*result__snippet[^"]*"[^>]*>([^<]+)<`),
		regexp.MustCompile(`<span[^>]*class="[^"]*snippet[^"]*"[^>]*>([^<]+)</span>`),
	}

	// Collect links
	var allLinks [][]string
	seenURLs := make(map[string]bool)

	for _, pattern := range linkPatterns {
		matches := pattern.FindAllStringSubmatch(html, maxResults*3)
		for _, match := range matches {
			if len(match) >= 3 {
				rawURL := match[1]
				if !seenURLs[rawURL] {
					seenURLs[rawURL] = true
					allLinks = append(allLinks, match)
				}
			}
		}
	}

	// Collect snippets
	var snippets []string
	for _, pattern := range snippetPatterns {
		matches := pattern.FindAllStringSubmatch(html, maxResults*3)
		for _, match := range matches {
			if len(match) >= 2 {
				snippets = append(snippets, match[1])
			}
		}
	}

	slog.Debug("DuckDuckGo HTML parsing", "linksFound", len(allLinks), "snippetsFound", len(snippets))

	snippetIdx := 0
	for _, match := range allLinks {
		if len(results) >= maxResults {
			break
		}

		rawURL := match[1]
		title := strings.TrimSpace(match[2])

		// Skip navigation and empty
		if title == "" || title == "..." || strings.Contains(title, "DuckDuckGo") {
			continue
		}

		// Decode the DuckDuckGo redirect URL
		actualURL := cleanDDGURL(rawURL)
		if actualURL == "" || !strings.HasPrefix(actualURL, "http") {
			continue
		}

		// Skip DuckDuckGo pages
		if strings.Contains(actualURL, "duckduckgo.com") {
			continue
		}

		// Get snippet if available
		snippet := ""
		if snippetIdx < len(snippets) {
			snippet = cleanHTMLText(snippets[snippetIdx])
			snippetIdx++
		}

		results = append(results, WebSearchResult{
			Title:   cleanHTMLText(title),
			URL:     actualURL,
			Snippet: snippet,
			Source:  extractDomain(actualURL),
		})
	}

	slog.Debug("DuckDuckGo HTML results", "count", len(results))

	return results
}

// searchInstant uses the DuckDuckGo Instant Answer API
func (d *DuckDuckGoProvider) searchInstant(ctx context.Context, query string) ([]WebSearchResult, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "BusinessOS/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("instant answer API returned status %d", resp.StatusCode)
	}

	var ddgResp struct {
		Abstract       string `json:"Abstract"`
		AbstractText   string `json:"AbstractText"`
		AbstractSource string `json:"AbstractSource"`
		AbstractURL    string `json:"AbstractURL"`
		Heading        string `json:"Heading"`
		RelatedTopics  []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"RelatedTopics"`
		Results []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
		} `json:"Results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ddgResp); err != nil {
		return nil, err
	}

	var results []WebSearchResult

	// Add abstract if available
	if ddgResp.AbstractText != "" && ddgResp.AbstractURL != "" {
		results = append(results, WebSearchResult{
			Title:   ddgResp.Heading,
			URL:     ddgResp.AbstractURL,
			Snippet: ddgResp.AbstractText,
			Source:  ddgResp.AbstractSource,
		})
	}

	// Add related topics
	for _, topic := range ddgResp.RelatedTopics {
		if topic.FirstURL != "" && topic.Text != "" {
			results = append(results, WebSearchResult{
				Title:   extractTitleFromText(topic.Text),
				URL:     topic.FirstURL,
				Snippet: topic.Text,
				Source:  extractDomain(topic.FirstURL),
			})
		}
	}

	// Add direct results
	for _, result := range ddgResp.Results {
		if result.FirstURL != "" && result.Text != "" {
			results = append(results, WebSearchResult{
				Title:   extractTitleFromText(result.Text),
				URL:     result.FirstURL,
				Snippet: result.Text,
				Source:  extractDomain(result.FirstURL),
			})
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results from instant answer API")
	}

	return results, nil
}

// ============================================================================
// MULTI-PROVIDER SEARCH SERVICE
// ============================================================================

// WebSearchService is the main search service with multi-provider support
type WebSearchService struct {
	providers []SearchProvider
	primary   SearchProvider
	client    *http.Client
	timeout   time.Duration
	mu        sync.RWMutex
}

// NewWebSearchService creates a new web search service with providers based on config
func NewWebSearchService() *WebSearchService {
	cfg := config.AppConfig
	if cfg == nil {
		// Fallback if config not loaded
		return &WebSearchService{
			providers: []SearchProvider{NewDuckDuckGoProvider()},
			primary:   NewDuckDuckGoProvider(),
			client:    &http.Client{Timeout: 15 * time.Second},
			timeout:   15 * time.Second,
		}
	}

	return NewWebSearchServiceWithConfig(cfg)
}

// NewWebSearchServiceWithConfig creates a search service with explicit config
func NewWebSearchServiceWithConfig(cfg *config.Config) *WebSearchService {
	var providers []SearchProvider
	var primary SearchProvider

	// Add providers based on available API keys
	if cfg.HasBraveSearch() {
		provider := NewBraveSearchProvider(cfg.BraveSearchAPIKey)
		providers = append(providers, provider)
		if primary == nil {
			primary = provider
		}
	}

	if cfg.HasSerper() {
		provider := NewSerperProvider(cfg.SerperAPIKey)
		providers = append(providers, provider)
		if primary == nil {
			primary = provider
		}
	}

	if cfg.HasTavily() {
		provider := NewTavilyProvider(cfg.TavilyAPIKey)
		providers = append(providers, provider)
		if primary == nil {
			primary = provider
		}
	}

	// Always add DuckDuckGo as fallback
	ddg := NewDuckDuckGoProvider()
	providers = append(providers, ddg)
	if primary == nil {
		primary = ddg
	}

	// Override primary if explicitly set in config
	if cfg.SearchProvider != "" && cfg.SearchProvider != "auto" {
		for _, p := range providers {
			if p.Name() == cfg.SearchProvider && p.Available() {
				primary = p
				break
			}
		}
	}

	slog.Info("Web search service initialized",
		"primary", primary.Name(),
		"providers", len(providers),
		"available", getAvailableProviderNames(providers))

	return &WebSearchService{
		providers: providers,
		primary:   primary,
		client:    &http.Client{Timeout: 15 * time.Second},
		timeout:   15 * time.Second,
	}
}

// getAvailableProviderNames returns names of available providers
func getAvailableProviderNames(providers []SearchProvider) []string {
	names := make([]string, 0, len(providers))
	for _, p := range providers {
		if p.Available() {
			names = append(names, p.Name())
		}
	}
	return names
}

// Search performs a web search with automatic fallback
func (s *WebSearchService) Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error) {
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 20 {
		maxResults = 20
	}

	s.mu.RLock()
	primary := s.primary
	providers := s.providers
	s.mu.RUnlock()

	// Try primary provider first
	if primary != nil && primary.Available() {
		response, err := primary.Search(ctx, query, maxResults)
		if err == nil && len(response.Results) > 0 {
			return response, nil
		}
		slog.Warn("Primary search provider failed, trying fallbacks",
			"provider", primary.Name(),
			"error", err)
	}

	// Try fallback providers
	var lastErr error
	for _, provider := range providers {
		if provider == primary || !provider.Available() {
			continue
		}

		response, err := provider.Search(ctx, query, maxResults)
		if err == nil && len(response.Results) > 0 {
			slog.Info("Fallback search provider succeeded", "provider", provider.Name())
			return response, nil
		}
		lastErr = err
		slog.Warn("Fallback search provider failed",
			"provider", provider.Name(),
			"error", err)
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all search providers failed: %w", lastErr)
	}
	return nil, fmt.Errorf("no search providers available")
}

// SearchWithProvider performs a search using a specific provider
func (s *WebSearchService) SearchWithProvider(ctx context.Context, providerName string, query string, maxResults int) (*WebSearchResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, provider := range s.providers {
		if provider.Name() == providerName && provider.Available() {
			return provider.Search(ctx, query, maxResults)
		}
	}

	return nil, fmt.Errorf("provider %q not found or not available", providerName)
}

// GetAvailableProviders returns list of available provider names
func (s *WebSearchService) GetAvailableProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return getAvailableProviderNames(s.providers)
}

// GetPrimaryProvider returns the name of the primary provider
func (s *WebSearchService) GetPrimaryProvider() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.primary != nil {
		return s.primary.Name()
	}
	return "none"
}

// ============================================================================
// QUERY OPTIMIZER
// ============================================================================

// QueryOptimizer transforms conversational input into optimized search queries
type QueryOptimizer struct{}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer() *QueryOptimizer {
	return &QueryOptimizer{}
}

// stopWords are common words to remove from queries
var stopWords = map[string]bool{
	"a": true, "an": true, "the": true, "is": true, "are": true, "was": true, "were": true,
	"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
	"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
	"should": true, "may": true, "might": true, "must": true, "shall": true,
	"i": true, "me": true, "my": true, "myself": true, "we": true, "our": true,
	"you": true, "your": true, "yourself": true, "he": true, "she": true, "it": true,
	"they": true, "them": true, "their": true, "this": true, "that": true, "these": true,
	"what": true, "which": true, "who": true, "whom": true, "where": true, "when": true,
	"why": true, "how": true, "can": true, "please": true, "help": true, "need": true,
	"want": true, "like": true, "know": true, "think": true, "just": true, "also": true,
	"about": true, "with": true, "from": true, "into": true, "some": true, "any": true,
	"tell": true, "explain": true, "describe": true, "give": true, "show": true,
}

// questionPrefixes are common question patterns to strip
var questionPrefixes = []string{
	"can you tell me",
	"could you explain",
	"what is the",
	"what are the",
	"how do i",
	"how can i",
	"how to",
	"why is",
	"why are",
	"why do",
	"where can i find",
	"where is",
	"when did",
	"when was",
	"who is",
	"who are",
	"tell me about",
	"explain to me",
	"i want to know",
	"i need to know",
	"i'm looking for",
	"looking for",
	"searching for",
	"find me",
	"help me with",
	"help me understand",
}

// OptimizeQuery transforms conversational input into a search-optimized query
func (o *QueryOptimizer) OptimizeQuery(input string) string {
	if input == "" {
		return ""
	}

	// Lowercase for processing
	query := strings.ToLower(strings.TrimSpace(input))

	// Remove question prefixes
	for _, prefix := range questionPrefixes {
		if strings.HasPrefix(query, prefix) {
			query = strings.TrimPrefix(query, prefix)
			query = strings.TrimSpace(query)
			break
		}
	}

	// Remove punctuation at the end
	query = strings.TrimRight(query, "?.!,;:")

	// Split into words and filter
	words := strings.Fields(query)
	var filteredWords []string

	for _, word := range words {
		// Clean the word
		cleanWord := strings.Trim(word, ".,!?;:'\"()[]{}")

		// Skip stop words unless it's a short query
		if len(words) > 3 && stopWords[cleanWord] {
			continue
		}

		// Skip very short words unless they're likely important (numbers, acronyms)
		if len(cleanWord) < 2 && !isNumeric(cleanWord) {
			continue
		}

		filteredWords = append(filteredWords, cleanWord)
	}

	// Reconstruct the query
	optimized := strings.Join(filteredWords, " ")

	// If we filtered too aggressively, use original minus question prefix
	if len(optimized) < 3 && len(input) > 3 {
		optimized = strings.TrimRight(query, "?.!,;:")
	}

	return optimized
}

// QueryType represents the type of search query
type QueryType int

const (
	QueryTypeGeneral QueryType = iota
	QueryTypeHowTo
	QueryTypeComparison
	QueryTypeNews
	QueryTypeResearch
	QueryTypeTechnical
)

// QueryContext provides additional context for query optimization
type QueryContext struct {
	QueryType       QueryType
	RequireRecent   bool
	PreferredDomain string
	Language        string
}

// DetectQueryType analyzes the input to determine the type of query
func (o *QueryOptimizer) DetectQueryType(input string) QueryType {
	lower := strings.ToLower(input)

	// Check for how-to patterns
	if strings.Contains(lower, "how to") || strings.Contains(lower, "how do i") ||
		strings.Contains(lower, "how can i") || strings.Contains(lower, "steps to") {
		return QueryTypeHowTo
	}

	// Check for comparison patterns
	if strings.Contains(lower, " vs ") || strings.Contains(lower, "versus") ||
		strings.Contains(lower, "compared to") || strings.Contains(lower, "difference between") ||
		strings.Contains(lower, "better than") {
		return QueryTypeComparison
	}

	// Check for news patterns
	if strings.Contains(lower, "latest") || strings.Contains(lower, "recent") ||
		strings.Contains(lower, "news") || strings.Contains(lower, "today") ||
		strings.Contains(lower, "announced") || strings.Contains(lower, "released") {
		return QueryTypeNews
	}

	// Check for research patterns
	if strings.Contains(lower, "research") || strings.Contains(lower, "study") ||
		strings.Contains(lower, "analysis") || strings.Contains(lower, "statistics") ||
		strings.Contains(lower, "data on") {
		return QueryTypeResearch
	}

	// Check for technical patterns
	if strings.Contains(lower, "error") || strings.Contains(lower, "code") ||
		strings.Contains(lower, "programming") || strings.Contains(lower, "api") ||
		strings.Contains(lower, "documentation") || strings.Contains(lower, "bug") ||
		strings.Contains(lower, "implement") {
		return QueryTypeTechnical
	}

	return QueryTypeGeneral
}

// OptimizeWithContext adds contextual modifiers based on query intent
func (o *QueryOptimizer) OptimizeWithContext(input string, context QueryContext) string {
	base := o.OptimizeQuery(input)
	if base == "" {
		return ""
	}

	var modifiers []string

	// Add recency modifier for current events
	if context.RequireRecent {
		modifiers = append(modifiers, "2024 2025")
	}

	// Add type-specific modifiers
	switch context.QueryType {
	case QueryTypeTechnical:
		modifiers = append(modifiers, "documentation tutorial guide")
	case QueryTypeNews:
		modifiers = append(modifiers, "news latest")
	case QueryTypeResearch:
		modifiers = append(modifiers, "research study analysis")
	case QueryTypeHowTo:
		modifiers = append(modifiers, "how to guide steps")
	case QueryTypeComparison:
		modifiers = append(modifiers, "vs comparison difference")
	}

	// Add domain restriction if specified
	if context.PreferredDomain != "" {
		modifiers = append(modifiers, "site:"+context.PreferredDomain)
	}

	if len(modifiers) > 0 {
		return base + " " + strings.Join(modifiers, " ")
	}

	return base
}

// GenerateSearchQueries creates multiple search queries for comprehensive results
func (o *QueryOptimizer) GenerateSearchQueries(input string, maxQueries int) []string {
	if maxQueries <= 0 {
		maxQueries = 3
	}

	queries := make([]string, 0, maxQueries)

	// Primary optimized query
	primary := o.OptimizeQuery(input)
	if primary != "" {
		queries = append(queries, primary)
	}

	// Detect query type and add contextual variant
	queryType := o.DetectQueryType(input)
	ctx := QueryContext{QueryType: queryType}

	contextual := o.OptimizeWithContext(input, ctx)
	if contextual != primary && contextual != "" {
		queries = append(queries, contextual)
	}

	// Add a more specific variant if query is long enough
	if len(queries) < maxQueries {
		words := strings.Fields(primary)
		if len(words) > 4 {
			// Take the most important words (skip first word if it's common)
			start := 0
			if stopWords[strings.ToLower(words[0])] {
				start = 1
			}
			end := min(start+4, len(words))
			specific := strings.Join(words[start:end], " ")
			if specific != primary {
				queries = append(queries, specific)
			}
		}
	}

	return queries
}

// SearchWithOptimization performs a search with automatic query optimization
func (s *WebSearchService) SearchWithOptimization(ctx context.Context, input string, maxResults int) (*WebSearchResponse, error) {
	optimizer := NewQueryOptimizer()
	optimizedQuery := optimizer.OptimizeQuery(input)

	if optimizedQuery == "" {
		optimizedQuery = input
	}

	return s.Search(ctx, optimizedQuery, maxResults)
}

// FormatResultsAsContext formats search results for injection into AI context
func (s *WebSearchService) FormatResultsAsContext(results *WebSearchResponse) string {
	if results == nil || len(results.Results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Web Search Results for: %q\n\n", results.Query))

	for i, result := range results.Results {
		sb.WriteString(fmt.Sprintf("### [%d] %s\n", i+1, result.Title))
		sb.WriteString(fmt.Sprintf("**Source:** %s\n", result.Source))
		sb.WriteString(fmt.Sprintf("**URL:** %s\n", result.URL))
		if result.Snippet != "" {
			sb.WriteString(fmt.Sprintf("**Summary:** %s\n", result.Snippet))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("*Search completed in %.0fms via %s*\n", results.SearchTime, results.Provider))

	return sb.String()
}

// ============================================================================
// CACHED SEARCH SERVICE
// ============================================================================

// CachedWebSearchService wraps WebSearchService with caching capabilities
type CachedWebSearchService struct {
	*WebSearchService
	pool         *pgxpool.Pool
	cacheTTL     time.Duration
	newsCacheTTL time.Duration
}

// NewCachedWebSearchService creates a new cached web search service
func NewCachedWebSearchService(pool *pgxpool.Pool) *CachedWebSearchService {
	return &CachedWebSearchService{
		WebSearchService: NewWebSearchService(),
		pool:             pool,
		cacheTTL:         1 * time.Hour,
		newsCacheTTL:     15 * time.Minute,
	}
}

// NewCachedWebSearchServiceWithConfig creates a cached search service with explicit config
func NewCachedWebSearchServiceWithConfig(pool *pgxpool.Pool, cfg *config.Config) *CachedWebSearchService {
	return &CachedWebSearchService{
		WebSearchService: NewWebSearchServiceWithConfig(cfg),
		pool:             pool,
		cacheTTL:         1 * time.Hour,
		newsCacheTTL:     15 * time.Minute,
	}
}

// hashQuery creates a SHA256 hash of the normalized query
func (s *CachedWebSearchService) hashQuery(query string) string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// cachedResult holds the structure of a cached search result
type cachedResult struct {
	ID            pgtype.UUID
	QueryHash     string
	OriginalQuery string
	Results       []byte
	ResultCount   int32
	Provider      string
	HitCount      int32
}

// SearchWithCache performs a search with caching
func (s *CachedWebSearchService) SearchWithCache(ctx context.Context, query string, maxResults int, userID string, conversationID *uuid.UUID) (*WebSearchResponse, error) {
	if s.pool == nil {
		// No pool available, fall back to direct search
		return s.Search(ctx, query, maxResults)
	}

	queryHash := s.hashQuery(query)

	// Try to get cached result
	cached, err := s.getCachedResult(ctx, queryHash, conversationID)
	if err == nil && cached != nil {
		slog.Debug("Web search cache hit", "queryHash", queryHash, "hitCount", cached.HitCount)

		// Increment hit count asynchronously
		go func() {
			_, _ = s.pool.Exec(context.Background(),
				"UPDATE web_search_results SET hit_count = hit_count + 1, last_hit_at = NOW() WHERE id = $1",
				cached.ID)
		}()

		// Parse cached results
		var results []WebSearchResult
		if err := json.Unmarshal(cached.Results, &results); err == nil {
			return &WebSearchResponse{
				Query:        cached.OriginalQuery,
				Results:      results,
				TotalResults: int(cached.ResultCount),
				SearchTime:   0, // Cached, no search time
				Provider:     cached.Provider,
				Cached:       true,
			}, nil
		}
	}

	// Cache miss - perform search
	response, err := s.SearchWithOptimization(ctx, query, maxResults)
	if err != nil {
		return nil, err
	}

	// Save to cache asynchronously
	go func() {
		s.saveToCache(context.Background(), query, response, userID, conversationID)
	}()

	return response, nil
}

// getCachedResult retrieves a cached result from the database
func (s *CachedWebSearchService) getCachedResult(ctx context.Context, queryHash string, conversationID *uuid.UUID) (*cachedResult, error) {
	var query string
	var args []interface{}

	if conversationID != nil {
		query = `SELECT id, query_hash, original_query, results, result_count, provider, hit_count
				 FROM web_search_results
				 WHERE query_hash = $1 AND conversation_id = $2 AND expires_at > NOW()
				 ORDER BY created_at DESC LIMIT 1`
		args = []interface{}{queryHash, *conversationID}
	} else {
		query = `SELECT id, query_hash, original_query, results, result_count, provider, hit_count
				 FROM web_search_results
				 WHERE query_hash = $1 AND expires_at > NOW()
				 ORDER BY created_at DESC LIMIT 1`
		args = []interface{}{queryHash}
	}

	row := s.pool.QueryRow(ctx, query, args...)

	var result cachedResult
	var provider *string
	err := row.Scan(&result.ID, &result.QueryHash, &result.OriginalQuery, &result.Results, &result.ResultCount, &provider, &result.HitCount)
	if err != nil {
		return nil, err
	}

	if provider != nil {
		result.Provider = *provider
	} else {
		result.Provider = "unknown"
	}

	return &result, nil
}

// saveToCache saves search results to the cache
func (s *CachedWebSearchService) saveToCache(ctx context.Context, originalQuery string, response *WebSearchResponse, userID string, conversationID *uuid.UUID) {
	queryHash := s.hashQuery(originalQuery)

	// Determine TTL based on query type
	optimizer := NewQueryOptimizer()
	queryType := optimizer.DetectQueryType(originalQuery)
	ttl := s.cacheTTL
	if queryType == QueryTypeNews {
		ttl = s.newsCacheTTL
	}

	// Serialize results
	resultsJSON, err := json.Marshal(response.Results)
	if err != nil {
		slog.Error("Failed to marshal search results for cache", "err", err)
		return
	}

	expiresAt := time.Now().Add(ttl)

	query := `INSERT INTO web_search_results (
		query_hash, original_query, optimized_query, user_id, conversation_id,
		results, result_count, provider, search_time_ms, expires_at
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	var userIDPtr *string
	if userID != "" {
		userIDPtr = &userID
	}

	_, err = s.pool.Exec(ctx, query,
		queryHash,
		originalQuery,
		response.Query, // optimized query
		userIDPtr,
		conversationID,
		resultsJSON,
		response.TotalResults,
		response.Provider,
		response.SearchTime,
		expiresAt,
	)

	if err != nil {
		slog.Error("Failed to save search result to cache", "err", err)
	} else {
		slog.Debug("Saved search result to cache", "queryHash", queryHash, "provider", response.Provider, "ttl", ttl)
	}
}

// CleanupCache removes expired cache entries
func (s *CachedWebSearchService) CleanupCache(ctx context.Context) (int64, error) {
	if s.pool == nil {
		return 0, nil
	}

	result, err := s.pool.Exec(ctx, "DELETE FROM web_search_results WHERE expires_at < NOW()")
	if err != nil {
		return 0, err
	}

	return result.RowsAffected(), nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// getRandomUserAgent returns a random user agent from the list
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// extractDomain extracts the domain from a URL
func extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Host
}

// cleanDDGURL decodes DuckDuckGo's redirect URL
func cleanDDGURL(ddgURL string) string {
	// DuckDuckGo URLs are in format: //duckduckgo.com/l/?uddg=ENCODED_URL&...
	if strings.Contains(ddgURL, "uddg=") {
		parsed, err := url.Parse(ddgURL)
		if err != nil {
			return ddgURL
		}
		uddg := parsed.Query().Get("uddg")
		if uddg != "" {
			decoded, err := url.QueryUnescape(uddg)
			if err == nil {
				return decoded
			}
		}
	}

	// Handle direct URLs
	if strings.HasPrefix(ddgURL, "http") {
		return ddgURL
	}
	if strings.HasPrefix(ddgURL, "//") {
		return "https:" + ddgURL
	}

	return ddgURL
}

// cleanHTMLText removes HTML tags and decodes entities
func cleanHTMLText(text string) string {
	// Remove HTML tags
	tagPattern := regexp.MustCompile(`<[^>]+>`)
	text = tagPattern.ReplaceAllString(text, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&#x27;", "'")
	text = strings.ReplaceAll(text, "&#x2F;", "/")

	return strings.TrimSpace(text)
}

// extractTitleFromText extracts a title from DuckDuckGo's text format
func extractTitleFromText(text string) string {
	// DuckDuckGo often has format "Title - Description"
	if idx := strings.Index(text, " - "); idx > 0 && idx < 100 {
		return text[:idx]
	}
	// Truncate if too long
	if len(text) > 100 {
		return text[:97] + "..."
	}
	return text
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
