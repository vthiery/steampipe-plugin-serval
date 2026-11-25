package serval

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"golang.org/x/time/rate"
)

const (
	defaultBaseURL = "https://public.api.serval.com"
	maxRetries     = 8
)

// limiter is shared by every request the plugin process makes. Serval does not
// document a rate limit or send rate-limit headers; 10 req/s keeps full scans
// (one access-membership call per user) reasonable without hammering the API.
var limiter = rate.NewLimiter(rate.Limit(10), 10)

// Row is a single API object. Columns read fields with transform.FromField("camelCaseKey").
type Row = map[string]interface{}

// Client wraps the Serval public REST API (v2).
type Client struct {
	clientID     string
	clientSecret string
	baseURL      string
	teamIDs      []string
	httpClient   *http.Client
}

type tokenEntry struct {
	token   string
	expires time.Time
}

var (
	tokenMu    sync.Mutex
	tokenCache = map[string]tokenEntry{}

	teamMu    sync.Mutex
	teamCache = map[string][]string{}
)

// errNotFound is returned for 404 responses.
var errNotFound = errors.New("not found")

// apiError is a non-retryable error response.
type apiError struct {
	Method string
	Path   string
	Status int
	Body   string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%s %s: status %d: %s", e.Method, e.Path, e.Status, e.Body)
}

// isForeignWorkspace reports whether the API refused a filter because the object it names
// (an app instance, resource, ...) lives in another workspace. Workspace-scoped tables
// scan every workspace, so such a filter only matches in one of them.
func isForeignWorkspace(err error) bool {
	var ae *apiError
	return errors.As(err, &ae) && ae.Status == http.StatusBadRequest &&
		strings.Contains(ae.Body, "does not belong to the specified workspace")
}

func getClient(_ context.Context, d *plugin.QueryData) (*Client, error) {
	cfg := GetConfig(d.Connection)
	pick := func(v *string, env string) string {
		if v != nil && *v != "" {
			return *v
		}
		return os.Getenv(env)
	}
	c := &Client{
		clientID:     pick(cfg.ClientID, "SERVAL_CLIENT_ID"),
		clientSecret: pick(cfg.ClientSecret, "SERVAL_CLIENT_SECRET"),
		baseURL:      strings.TrimRight(pick(cfg.BaseURL, "SERVAL_BASE_URL"), "/"),
		teamIDs:      cfg.TeamIDs,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
	if c.clientID == "" || c.clientSecret == "" {
		return nil, errors.New("client_id and client_secret must be configured (or set SERVAL_CLIENT_ID and SERVAL_CLIENT_SECRET)")
	}
	if c.baseURL == "" {
		c.baseURL = defaultBaseURL
	}
	if len(c.teamIDs) == 0 {
		for _, id := range strings.Split(os.Getenv("SERVAL_TEAM_IDS"), ",") {
			if id = strings.TrimSpace(id); id != "" {
				c.teamIDs = append(c.teamIDs, id)
			}
		}
	}
	return c, nil
}

// token returns a cached OAuth access token, fetching a new one when needed.
func (c *Client) token(ctx context.Context, refresh bool) (string, error) {
	key := c.baseURL + "|" + c.clientID
	tokenMu.Lock()
	defer tokenMu.Unlock()
	if e, ok := tokenCache[key]; ok && !refresh && time.Now().Before(e.expires) {
		return e.token, nil
	}
	form := url.Values{"grant_type": {"client_credentials"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v2/auth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(c.clientID, c.clientSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("requesting token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("requesting token from %s: status %d: %s (check client_id, client_secret and base_url: tenants are regional)", c.baseURL, resp.StatusCode, truncate(body))
	}
	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(body, &tok); err != nil || tok.AccessToken == "" {
		return "", fmt.Errorf("decoding token response: %v", err)
	}
	ttl := time.Duration(tok.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = time.Hour
	}
	// Refresh a minute early so a long scan never sends an expired token.
	tokenCache[key] = tokenEntry{token: tok.AccessToken, expires: time.Now().Add(ttl - time.Minute)}
	return tok.AccessToken, nil
}

// do performs an authenticated request, retrying on 429 and 5xx, and decodes the body into result.
func (c *Client) do(ctx context.Context, method, path string, params url.Values, body interface{}, result interface{}) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	var payload []byte
	if body != nil {
		var err error
		if payload, err = json.Marshal(body); err != nil {
			return err
		}
	}

	plugin.Logger(ctx).Debug("serval.request", "method", method, "url", u)
	var lastErr error
	refreshed := false
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
		tok, err := c.token(ctx, false)
		if err != nil {
			return err
		}
		var rd io.Reader
		if payload != nil {
			rd = bytes.NewReader(payload)
		}
		req, err := http.NewRequestWithContext(ctx, method, u, rd)
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Accept", "application/json")
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
		} else {
			data, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			switch {
			case readErr != nil:
				lastErr = fmt.Errorf("reading response body: %w", readErr)
			case resp.StatusCode == http.StatusOK:
				if err := json.Unmarshal(data, result); err != nil {
					return fmt.Errorf("unmarshalling response from %s: %w", path, err)
				}
				return nil
			case resp.StatusCode == http.StatusNotFound:
				return errNotFound
			case resp.StatusCode == http.StatusUnauthorized && !refreshed:
				// Token revoked or expired early: fetch a new one and retry once.
				refreshed = true
				if _, err := c.token(ctx, true); err != nil {
					return err
				}
				continue
			case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
				lastErr = fmt.Errorf("%s %s: status %d: %s", method, path, resp.StatusCode, truncate(data))
				if s, err := strconv.Atoi(resp.Header.Get("Retry-After")); err == nil && s > 0 {
					if err := sleep(ctx, time.Duration(s)*time.Second); err != nil {
						return err
					}
					continue
				}
			default:
				return &apiError{Method: method, Path: path, Status: resp.StatusCode, Body: truncate(data)}
			}
		}
		// 0.5s, 1s, 2s, ... capped at 30s.
		backoff := time.Duration(1<<attempt) * 500 * time.Millisecond
		if backoff > 30*time.Second {
			backoff = 30 * time.Second
		}
		if err := sleep(ctx, backoff); err != nil {
			return err
		}
	}
	return lastErr
}

type page struct {
	Data          []Row  `json:"data"`
	NextPageToken string `json:"nextPageToken"`
}

// paginate walks a token-paginated endpoint and calls fn for every row until fn returns
// false. GET requests carry pageSize/pageToken as query parameters, POST requests in the
// JSON body.
func (c *Client) paginate(ctx context.Context, method, path string, params url.Values, body Row, pageSize int, fn func(Row) bool) error {
	if params == nil {
		params = url.Values{}
	}
	if body == nil && method == http.MethodPost {
		body = Row{}
	}
	token := ""
	for {
		if method == http.MethodGet {
			params.Set("pageSize", strconv.Itoa(pageSize))
			if token != "" {
				params.Set("pageToken", token)
			}
		} else {
			body["pageSize"] = pageSize
			if token != "" {
				body["pageToken"] = token
			}
		}
		var p page
		var reqBody interface{}
		if method == http.MethodPost {
			reqBody = body
		}
		if err := c.do(ctx, method, path, params, reqBody, &p); err != nil {
			return err
		}
		for _, row := range p.Data {
			if !fn(row) {
				return nil
			}
		}
		if p.NextPageToken == "" || p.NextPageToken == token || len(p.Data) == 0 {
			return nil
		}
		token = p.NextPageToken
	}
}

// getOne fetches a single object by path, returning nil (no row) on 404.
func (c *Client) getOne(ctx context.Context, path string) (Row, error) {
	var res struct {
		Data Row `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, path, nil, nil, &res); err != nil {
		if errors.Is(err, errNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return res.Data, nil
}

// teams returns the workspace IDs to query: the configured ones, or every workspace
// the key can list.
func (c *Client) teams(ctx context.Context) ([]string, error) {
	if len(c.teamIDs) > 0 {
		return c.teamIDs, nil
	}
	key := c.baseURL + "|" + c.clientID
	teamMu.Lock()
	defer teamMu.Unlock()
	if ids, ok := teamCache[key]; ok {
		return ids, nil
	}
	var ids []string
	err := c.paginate(ctx, http.MethodGet, "/v2/teams", nil, nil, 100, func(r Row) bool {
		if id, ok := r["id"].(string); ok {
			ids = append(ids, id)
		}
		return true
	})
	if err != nil {
		return nil, fmt.Errorf("listing workspaces (set team_ids in the connection config if the API key cannot list teams): %w", err)
	}
	teamCache[key] = ids
	return ids, nil
}

func sleep(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

func truncate(b []byte) string {
	if len(b) > 300 {
		return string(b[:300]) + "…"
	}
	return string(b)
}
