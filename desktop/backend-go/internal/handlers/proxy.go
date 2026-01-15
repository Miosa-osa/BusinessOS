package handlers

import (
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ProxyHandler handles proxying requests to external URLs, stripping iframe-blocking headers
// This allows embedding external web apps within BusinessOS iframes

// HandleProxyURL proxies a GET request to an external URL
// Route: GET /api/proxy?url=<encoded-url>
func (h *Handlers) HandleProxyURL(c *gin.Context) {
	targetURL := c.Query("url")
	if targetURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url parameter is required"})
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL"})
		return
	}

	// Only allow http/https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only http/https URLs are allowed"})
		return
	}

	slog.Info("Proxying request",
		"url", targetURL,
		"host", parsedURL.Host,
	)

	// Create HTTP client with reasonable timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
		// Don't follow redirects automatically - let us handle them
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// Create the proxy request
	proxyReq, err := http.NewRequestWithContext(c.Request.Context(), "GET", targetURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// Copy some headers from the original request
	if accept := c.GetHeader("Accept"); accept != "" {
		proxyReq.Header.Set("Accept", accept)
	} else {
		proxyReq.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	}

	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		proxyReq.Header.Set("Accept-Language", acceptLang)
	} else {
		proxyReq.Header.Set("Accept-Language", "en-US,en;q=0.5")
	}

	// Set a realistic User-Agent
	proxyReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Accept compressed responses
	proxyReq.Header.Set("Accept-Encoding", "gzip, deflate")

	// Execute the request
	resp, err := client.Do(proxyReq)
	if err != nil {
		slog.Error("Proxy request failed", "error", err, "url", targetURL)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "failed to fetch URL",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Handle redirects
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		location := resp.Header.Get("Location")
		if location != "" {
			// Return redirect info to client so it can request the new URL
			c.JSON(http.StatusTemporaryRedirect, gin.H{
				"redirect": true,
				"location": location,
			})
			return
		}
	}

	// Copy allowed response headers (excluding iframe-blocking ones)
	allowedHeaders := []string{
		"Content-Type",
		"Content-Language",
		"Cache-Control",
		"ETag",
		"Last-Modified",
		"Expires",
		"Date",
	}

	for _, header := range allowedHeaders {
		if val := resp.Header.Get(header); val != "" {
			c.Header(header, val)
		}
	}

	// CRITICAL: We explicitly DO NOT copy these headers:
	// - X-Frame-Options (blocks iframe embedding)
	// - Content-Security-Policy (may block iframe embedding with frame-ancestors)
	// - X-Content-Type-Options (can cause issues)

	// Set permissive CSP that allows embedding
	c.Header("X-Frame-Options", "ALLOWALL")
	c.Header("Content-Security-Policy", "frame-ancestors *")

	// Handle gzipped responses
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			slog.Error("Failed to decompress gzip response", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to decompress response"})
			return
		}
		defer gzReader.Close()
		reader = gzReader
		// Remove Content-Encoding since we're decompressing
		c.Header("Content-Encoding", "")
	}

	// Read the body
	body, err := io.ReadAll(reader)
	if err != nil {
		slog.Error("Failed to read response body", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read response"})
		return
	}

	// For HTML responses, rewrite relative URLs to absolute
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		body = rewriteHTMLURLs(body, parsedURL)
	}

	// Set the status code
	c.Status(resp.StatusCode)

	// Write the body
	c.Writer.Write(body)
}

// rewriteHTMLURLs rewrites relative URLs in HTML to absolute URLs
// This is necessary because the proxied page will think it's on our domain
func rewriteHTMLURLs(body []byte, baseURL *url.URL) []byte {
	html := string(body)

	// Base URL without path
	baseOrigin := fmt.Sprintf("%s://%s", baseURL.Scheme, baseURL.Host)

	// Inject a <base> tag at the start of <head> to handle relative URLs
	// This is the most reliable way to fix relative URLs
	baseTag := fmt.Sprintf(`<base href="%s/">`, baseOrigin)

	// Try to inject after <head>
	if strings.Contains(html, "<head>") {
		html = strings.Replace(html, "<head>", "<head>"+baseTag, 1)
	} else if strings.Contains(html, "<HEAD>") {
		html = strings.Replace(html, "<HEAD>", "<HEAD>"+baseTag, 1)
	} else if strings.Contains(html, "<html>") {
		// Fallback: inject after <html>
		html = strings.Replace(html, "<html>", "<html><head>"+baseTag+"</head>", 1)
	} else if strings.Contains(html, "<HTML>") {
		html = strings.Replace(html, "<HTML>", "<HTML><head>"+baseTag+"</head>", 1)
	}

	// Also inject a script that handles dynamic requests
	// This helps with JavaScript that makes fetch/XHR calls
	proxyScript := `<script>
(function() {
	// Store original fetch
	var originalFetch = window.fetch;

	// Override fetch to handle CORS
	window.fetch = function(url, options) {
		// If it's a relative URL, make it absolute
		if (url && typeof url === 'string' && !url.startsWith('http')) {
			url = '` + baseOrigin + `' + (url.startsWith('/') ? '' : '/') + url;
		}
		return originalFetch.call(this, url, options);
	};

	// Override XMLHttpRequest open
	var originalXHROpen = XMLHttpRequest.prototype.open;
	XMLHttpRequest.prototype.open = function(method, url, async, user, password) {
		if (url && typeof url === 'string' && !url.startsWith('http')) {
			url = '` + baseOrigin + `' + (url.startsWith('/') ? '' : '/') + url;
		}
		return originalXHROpen.call(this, method, url, async, user, password);
	};
})();
</script>`

	// Inject the script after <head> or at the start
	if strings.Contains(html, "</head>") {
		html = strings.Replace(html, "</head>", proxyScript+"</head>", 1)
	} else if strings.Contains(html, "</HEAD>") {
		html = strings.Replace(html, "</HEAD>", proxyScript+"</HEAD>", 1)
	}

	return []byte(html)
}

// HandleProxyPost handles POST requests to external URLs
// Route: POST /api/proxy
func (h *Handlers) HandleProxyPost(c *gin.Context) {
	var req struct {
		URL     string            `json:"url" binding:"required"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    string            `json:"body"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to GET if no method specified
	if req.Method == "" {
		req.Method = "GET"
	}

	// Validate URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid URL"})
		return
	}

	slog.Info("Proxying POST request",
		"url", req.URL,
		"method", req.Method,
	)

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create the request
	var bodyReader io.Reader
	if req.Body != "" {
		bodyReader = strings.NewReader(req.Body)
	}

	proxyReq, err := http.NewRequestWithContext(c.Request.Context(), req.Method, req.URL, bodyReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}

	// Set headers
	for key, val := range req.Headers {
		proxyReq.Header.Set(key, val)
	}

	// Default headers
	if proxyReq.Header.Get("User-Agent") == "" {
		proxyReq.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")
	}

	// Execute
	resp, err := client.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":   "request failed",
			"details": err.Error(),
		})
		return
	}
	defer resp.Body.Close()

	// Read body
	body, _ := io.ReadAll(resp.Body)

	// Return response info
	c.JSON(http.StatusOK, gin.H{
		"status":  resp.StatusCode,
		"headers": resp.Header,
		"body":    string(body),
	})
}
