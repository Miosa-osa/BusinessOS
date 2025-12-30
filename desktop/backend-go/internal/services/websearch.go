package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// userAgents is a list of common browser user agents for rotation
var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Edg/120.0.0.0",
}

// getRandomUserAgent returns a random user agent from the list
func getRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// WebSearchResult represents a single search result
type WebSearchResult struct {
	Title       string `json:"title"`
	URL         string `json:"url"`
	Snippet     string `json:"snippet"`
	Source      string `json:"source"`
	PublishedAt string `json:"published_at,omitempty"`
}

// WebSearchResponse represents the full search response
type WebSearchResponse struct {
	Query        string            `json:"query"`
	Results      []WebSearchResult `json:"results"`
	TotalResults int               `json:"total_results"`
	SearchTime   float64           `json:"search_time_ms"`
	Provider     string            `json:"provider"`
}

// WebSearchService handles web search operations
type WebSearchService struct {
	client  *http.Client
	timeout time.Duration
}

// NewWebSearchService creates a new web search service
func NewWebSearchService() *WebSearchService {
	return &WebSearchService{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		timeout: 15 * time.Second,
	}
}

// Search performs a web search using DuckDuckGo
func (s *WebSearchService) Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error) {
	startTime := time.Now()

	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 20 {
		maxResults = 20
	}

	// Try DuckDuckGo HTML search first
	results, err := s.searchDuckDuckGoHTML(ctx, query, maxResults)
	if err != nil {
		// Fallback to instant answers API
		results, err = s.searchDuckDuckGoInstant(ctx, query)
		if err != nil {
			return nil, fmt.Errorf("web search failed: %w", err)
		}
	}

	return &WebSearchResponse{
		Query:        query,
		Results:      results,
		TotalResults: len(results),
		SearchTime:   float64(time.Since(startTime).Milliseconds()),
		Provider:     "duckduckgo",
	}, nil
}

// searchDuckDuckGoHTML searches using DuckDuckGo's HTML endpoint
func (s *WebSearchService) searchDuckDuckGoHTML(ctx context.Context, query string, maxResults int) ([]WebSearchResult, error) {
	// DuckDuckGo HTML search endpoint
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// Set headers to mimic a browser with random User-Agent
	req.Header.Set("User-Agent", getRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.client.Do(req)
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

	return s.parseDuckDuckGoHTML(string(body), maxResults), nil
}

// parseDuckDuckGoHTML parses the HTML response from DuckDuckGo
func (s *WebSearchService) parseDuckDuckGoHTML(html string, maxResults int) []WebSearchResult {
	var results []WebSearchResult

	// Pattern to match search results
	// DuckDuckGo HTML results are in <div class="result"> blocks
	resultPattern := regexp.MustCompile(`<a rel="nofollow" class="result__a" href="([^"]+)"[^>]*>([^<]+)</a>`)
	snippetPattern := regexp.MustCompile(`<a class="result__snippet"[^>]*>([^<]+(?:<[^>]+>[^<]*</[^>]+>)*[^<]*)</a>`)

	// Find all result links
	linkMatches := resultPattern.FindAllStringSubmatch(html, maxResults*2)
	snippetMatches := snippetPattern.FindAllStringSubmatch(html, maxResults*2)

	for i := 0; i < len(linkMatches) && len(results) < maxResults; i++ {
		if len(linkMatches[i]) < 3 {
			continue
		}

		rawURL := linkMatches[i][1]
		title := strings.TrimSpace(linkMatches[i][2])

		// Decode the DuckDuckGo redirect URL
		actualURL := s.decodeDDGURL(rawURL)
		if actualURL == "" {
			continue
		}

		// Get snippet if available
		snippet := ""
		if i < len(snippetMatches) && len(snippetMatches[i]) > 1 {
			snippet = s.cleanHTML(snippetMatches[i][1])
		}

		// Extract source domain
		source := s.extractDomain(actualURL)

		results = append(results, WebSearchResult{
			Title:   s.cleanHTML(title),
			URL:     actualURL,
			Snippet: snippet,
			Source:  source,
		})
	}

	return results
}

// decodeDDGURL decodes DuckDuckGo's redirect URL
func (s *WebSearchService) decodeDDGURL(ddgURL string) string {
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

// searchDuckDuckGoInstant uses the DuckDuckGo Instant Answer API
func (s *WebSearchService) searchDuckDuckGoInstant(ctx context.Context, query string) ([]WebSearchResult, error) {
	apiURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "BusinessOS/1.0")

	resp, err := s.client.Do(req)
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
			Text      string `json:"Text"`
			FirstURL  string `json:"FirstURL"`
			Result    string `json:"Result"`
		} `json:"RelatedTopics"`
		Results []struct {
			Text     string `json:"Text"`
			FirstURL string `json:"FirstURL"`
			Result   string `json:"Result"`
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
				Title:   s.extractTitleFromText(topic.Text),
				URL:     topic.FirstURL,
				Snippet: topic.Text,
				Source:  s.extractDomain(topic.FirstURL),
			})
		}
	}

	// Add direct results
	for _, result := range ddgResp.Results {
		if result.FirstURL != "" && result.Text != "" {
			results = append(results, WebSearchResult{
				Title:   s.extractTitleFromText(result.Text),
				URL:     result.FirstURL,
				Snippet: result.Text,
				Source:  s.extractDomain(result.FirstURL),
			})
		}
	}

	return results, nil
}

// cleanHTML removes HTML tags and decodes entities
func (s *WebSearchService) cleanHTML(text string) string {
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

	return strings.TrimSpace(text)
}

// extractDomain extracts the domain from a URL
func (s *WebSearchService) extractDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Host
}

// extractTitleFromText extracts a title from DuckDuckGo's text format
func (s *WebSearchService) extractTitleFromText(text string) string {
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
