package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// HTTPResponse mirrors the Elixir response shape: %{status, headers, body}.
// Body is already JSON-decoded when ParseJSON is true and content-type includes
// "json"; otherwise it is the raw []byte.
type HTTPResponse struct {
	Status  int
	Headers http.Header
	Body    any
}

// HTTPOption configures a single request. Composable so adapters can build up
// headers + rate-limit + timeout in one place.
type HTTPOption func(*httpOpts)

type httpOpts struct {
	method    string
	headers   http.Header
	body      any
	timeout   time.Duration
	parseJSON bool
	bucket    *BucketSpec
}

// BucketSpec describes a rate-limit bucket the HTTP call should wait on.
// Rate is requests per second; Burst is the token-bucket size.
type BucketSpec struct {
	Key   string
	Rate  float64
	Burst int
}

// WithMethod sets the HTTP verb (default GET).
func WithMethod(m string) HTTPOption { return func(o *httpOpts) { o.method = m } }

// WithHeader adds a single header; repeated calls append.
func WithHeader(k, v string) HTTPOption {
	return func(o *httpOpts) {
		if o.headers == nil {
			o.headers = http.Header{}
		}
		o.headers.Add(k, v)
	}
}

// WithBearer is shorthand for setting an OAuth Authorization header.
func WithBearer(token string) HTTPOption {
	return WithHeader("Authorization", "Bearer "+token)
}

// WithJSONBody JSON-encodes the body and sets Content-Type.
func WithJSONBody(body any) HTTPOption {
	return func(o *httpOpts) {
		o.body = body
	}
}

// WithTimeout sets the request timeout (default 10s).
func WithTimeout(d time.Duration) HTTPOption { return func(o *httpOpts) { o.timeout = d } }

// WithoutJSONParse disables automatic JSON body decoding.
func WithoutJSONParse() HTTPOption { return func(o *httpOpts) { o.parseJSON = false } }

// WithRateLimit queues this request behind the named bucket before firing.
func WithRateLimit(key string, rate float64, burst int) HTTPOption {
	return func(o *httpOpts) { o.bucket = &BucketSpec{Key: key, Rate: rate, Burst: burst} }
}

// DoRequest is the canonical HTTP entry point for adapters.
// It handles JSON encode/decode, Retry-After extraction, and optional rate
// limiting. The returned error is nil for any successful round-trip, even
// on 4xx/5xx — callers inspect .Status themselves.
func DoRequest(ctx context.Context, url string, options ...HTTPOption) (*HTTPResponse, error) {
	opts := &httpOpts{
		method:    http.MethodGet,
		parseJSON: true,
		timeout:   10 * time.Second,
	}
	for _, o := range options {
		o(opts)
	}

	if opts.bucket != nil {
		if err := waitForBucket(ctx, opts.bucket); err != nil {
			return nil, fmt.Errorf("connectors/http: ratelimit: %w", err)
		}
	}

	var bodyReader io.Reader
	if opts.body != nil {
		payload, err := json.Marshal(opts.body)
		if err != nil {
			return nil, fmt.Errorf("connectors/http: marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(payload)
	}

	reqCtx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, opts.method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("connectors/http: new request: %w", err)
	}
	if opts.headers == nil {
		opts.headers = http.Header{}
	}
	if opts.headers.Get("Accept") == "" {
		opts.headers.Set("Accept", "application/json")
	}
	if opts.body != nil && opts.headers.Get("Content-Type") == "" {
		opts.headers.Set("Content-Type", "application/json")
	}
	req.Header = opts.headers

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connectors/http: request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("connectors/http: read body: %w", err)
	}

	out := &HTTPResponse{Status: resp.StatusCode, Headers: resp.Header}
	ct := strings.ToLower(resp.Header.Get("Content-Type"))
	if opts.parseJSON && strings.Contains(ct, "json") && len(raw) > 0 {
		var decoded any
		if err := json.Unmarshal(raw, &decoded); err == nil {
			out.Body = decoded
			return out, nil
		}
	}
	out.Body = raw
	return out, nil
}

// RetryAfter extracts the Retry-After header as a Duration. Returns 0 if
// the header is missing or malformed. Matches the Elixir behaviour of
// returning milliseconds, expressed here as time.Duration.
func RetryAfter(headers http.Header) time.Duration {
	v := headers.Get("Retry-After")
	if v == "" {
		return 0
	}
	if n, err := strconv.Atoi(v); err == nil && n > 0 {
		return time.Duration(n) * time.Second
	}
	return 0
}
