package utils

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// FaviconFetcher handles fetching favicons from URLs
type FaviconFetcher struct {
	client *http.Client
}

// NewFaviconFetcher creates a new favicon fetcher with timeout
func NewFaviconFetcher() *FaviconFetcher {
	return &FaviconFetcher{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// FetchFaviconURL attempts to find and return the favicon URL for a given website
func (f *FaviconFetcher) FetchFaviconURL(websiteURL string) (string, error) {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Ensure the URL has a scheme
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "https"
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	// Strategy 1: Try /favicon.ico
	faviconURL := baseURL + "/favicon.ico"
	if f.urlExists(faviconURL) {
		slog.Info("Found favicon via /favicon.ico", "url", faviconURL)
		return faviconURL, nil
	}

	// Strategy 2: Parse HTML and look for <link rel="icon"> tags
	faviconURL, err = f.fetchFromHTML(websiteURL)
	if err == nil && faviconURL != "" {
		// If relative URL, make it absolute
		if strings.HasPrefix(faviconURL, "/") {
			faviconURL = baseURL + faviconURL
		} else if !strings.HasPrefix(faviconURL, "http") {
			faviconURL = baseURL + "/" + faviconURL
		}
		slog.Info("Found favicon via HTML parsing", "url", faviconURL)
		return faviconURL, nil
	}

	// Strategy 3: Use Google's favicon service as fallback
	googleFaviconURL := fmt.Sprintf("https://www.google.com/s2/favicons?domain=%s&sz=128", parsedURL.Host)
	slog.Info("Using Google favicon service as fallback", "url", googleFaviconURL)
	return googleFaviconURL, nil
}

// urlExists checks if a URL returns a successful response
func (f *FaviconFetcher) urlExists(urlStr string) bool {
	resp, err := f.client.Head(urlStr)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// fetchFromHTML parses the HTML of a page to find favicon links
func (f *FaviconFetcher) fetchFromHTML(websiteURL string) (string, error) {
	resp, err := f.client.Get(websiteURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch HTML: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Limit reading to avoid memory issues
	limitedReader := io.LimitReader(resp.Body, 1024*1024) // 1MB limit
	doc, err := html.Parse(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Search for favicon link tags
	var faviconURL string
	var findFavicon func(*html.Node)
	findFavicon = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "link" {
			var rel, href string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "rel":
					rel = strings.ToLower(attr.Val)
				case "href":
					href = attr.Val
				}
			}
			// Look for icon, shortcut icon, apple-touch-icon
			if (strings.Contains(rel, "icon") || strings.Contains(rel, "shortcut")) && href != "" {
				if faviconURL == "" || strings.Contains(rel, "icon") {
					faviconURL = href
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findFavicon(c)
		}
	}

	findFavicon(doc)
	return faviconURL, nil
}
