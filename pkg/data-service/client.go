package data_service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	errors2 "github.com/controlplane-com/libs-go/pkg/errors"
	"github.com/controlplane-com/libs-go/pkg/logging"
)

const (
	rateLimitMaxAttempts    = 5
	rateLimitInitialBackoff = time.Second
	rateLimitMaxBackoff     = 30 * time.Second

	// RatelimitBypassHeader carries the token that exempts a request from
	// data-service rate limiting.
	RatelimitBypassHeader = "x-ratelimit-bypass"
)

type DataServiceClient struct {
	httpClient  *http.Client
	url         string
	authHeader  string
	serviceName string
	headers     http.Header
}

func NewClient(url string, token string, serviceName string) *DataServiceClient {
	return &DataServiceClient{
		authHeader: fmt.Sprintf("Bearer %s", token),
		url:        url,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
		serviceName: serviceName,
		headers:     http.Header{},
	}
}

func NewCustomClient(url string, token string, serviceName string, httpClient *http.Client) *DataServiceClient {
	if httpClient == nil {
		return NewClient(url, token, serviceName)
	}
	return &DataServiceClient{
		authHeader:  fmt.Sprintf("Bearer %s", token),
		url:         url,
		httpClient:  httpClient,
		serviceName: serviceName,
		headers:     http.Header{},
	}
}

// SetRatelimitBypass presents the bypass token so this client's calls skip
// data-service rate limiting. An empty token is a no-op.
func (c *DataServiceClient) SetRatelimitBypass(token string) {
	if token == "" {
		return
	}
	c.SetHeader(RatelimitBypassHeader, token)
}

func (c *DataServiceClient) SetHeader(key string, value string) {
	c.headers.Set(key, value)
}

func (c *DataServiceClient) DelHeader(key string) {
	c.headers.Del(key)
}

func (c *DataServiceClient) GetHeader(key string) string {
	return c.headers.Get(key)
}

func (c *DataServiceClient) SetToken(token string) {
	c.authHeader = fmt.Sprintf("Bearer %s", token)
}

func (c *DataServiceClient) Get(resource string, responseBodyOut any) (*http.Response, error) {
	request, err := c.Request(http.MethodGet, resource, "")
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

func (c *DataServiceClient) Post(resource string, requestBody string, responseBodyOut any) (*http.Response, error) {
	request, err := c.Request(http.MethodPost, resource, requestBody)
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

func (c *DataServiceClient) Patch(resource string, requestBody string, responseBodyOut any) (*http.Response, error) {
	request, err := c.Request(http.MethodPatch, resource, requestBody)
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

func (c *DataServiceClient) Put(resource string, requestBody string, responseBodyOut any) (*http.Response, error) {
	request, err := c.Request(http.MethodPut, resource, requestBody)
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

func (c *DataServiceClient) Delete(resource string) (*http.Response, error) {
	request, err := c.Request(http.MethodDelete, resource, "")
	if err != nil {
		return nil, err
	}
	return c.Do(request, nil)
}

func (c *DataServiceClient) Options(resource string) (*http.Response, error) {
	request, err := c.Request(http.MethodOptions, resource, "")
	if err != nil {
		return nil, err
	}
	return c.Do(request, nil)
}

func (c *DataServiceClient) addRequestHeaders(request *http.Request) {
	request.Header.Set("Authorization", c.authHeader)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", c.serviceName)
	for k := range c.headers {
		request.Header.Set(k, c.headers.Get(k))
	}
}

func (c *DataServiceClient) Request(method string, resource string, requestBody string) (*http.Request, error) {
	request, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.url, resource), strings.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	c.addRequestHeaders(request)
	return request, nil
}

func (c *DataServiceClient) Do(request *http.Request, responseBodyOut any) (*http.Response, error) {
	if !strings.HasPrefix(request.URL.String(), c.url) {
		return nil, errors.New("invalid url")
	}
	requestInfo := fmt.Sprintf("%s %s", request.Method, request.URL)
	var response *http.Response
	var body []byte
	for attempt := 1; ; attempt++ {
		var err error
		response, err = c.httpClient.Do(request)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", requestInfo, err)
		}
		body, err = io.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("%s: %w", requestInfo, err)
		}
		if response.StatusCode != http.StatusTooManyRequests || attempt >= rateLimitMaxAttempts {
			break
		}
		backoff := rateLimitBackoff(attempt, response.Header.Get("Retry-After"))
		logging.Logger().Sugar().Warnf("rate limited (429) on %s, retrying in %s (attempt %d/%d)", requestInfo, backoff, attempt, rateLimitMaxAttempts)
		select {
		case <-request.Context().Done():
			return nil, fmt.Errorf("%s: %w", requestInfo, request.Context().Err())
		case <-time.After(backoff):
		}
		if request.GetBody != nil {
			if request.Body, err = request.GetBody(); err != nil {
				return nil, fmt.Errorf("%s: %w", requestInfo, err)
			}
		}
	}
	if response.StatusCode > 299 {
		if response.StatusCode == 401 || response.StatusCode == 403 {
			return response, errors2.UnauthorizedError
		}
		return response, errors2.NewErrorCode(fmt.Sprintf("invalid HTTP response code (%d) from %s. Details: %s", response.StatusCode, requestInfo, string(body)), response.StatusCode)
	}
	if responseBodyOut != nil {
		err := json.Unmarshal(body, responseBodyOut)
		if err != nil {
			return response, err
		}
	}
	return response, nil
}

// rateLimitBackoff honors Retry-After seconds when present, else backs off exponentially.
func rateLimitBackoff(attempt int, retryAfter string) time.Duration {
	if retryAfter != "" {
		if secs, err := strconv.Atoi(retryAfter); err == nil && secs >= 0 {
			d := time.Duration(secs) * time.Second
			if d > rateLimitMaxBackoff {
				return rateLimitMaxBackoff
			}
			return d
		}
	}
	backoff := rateLimitInitialBackoff << (attempt - 1)
	if backoff > rateLimitMaxBackoff {
		return rateLimitMaxBackoff
	}
	return backoff
}

// RequestWithContext creates a request with context support.
func (c *DataServiceClient) RequestWithContext(ctx context.Context, method string, resource string, requestBody string) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, fmt.Sprintf("%s%s", c.url, resource), strings.NewReader(requestBody))
	if err != nil {
		return nil, err
	}
	c.addRequestHeaders(request)
	return request, nil
}

// GetWithContext performs a GET request with context support.
func (c *DataServiceClient) GetWithContext(ctx context.Context, resource string, responseBodyOut any) (*http.Response, error) {
	request, err := c.RequestWithContext(ctx, http.MethodGet, resource, "")
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

// PostWithContext performs a POST request with context support.
func (c *DataServiceClient) PostWithContext(ctx context.Context, resource string, requestBody string, responseBodyOut any) (*http.Response, error) {
	request, err := c.RequestWithContext(ctx, http.MethodPost, resource, requestBody)
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

// PatchWithContext performs a PATCH request with context support.
func (c *DataServiceClient) PatchWithContext(ctx context.Context, resource string, requestBody string, responseBodyOut any) (*http.Response, error) {
	request, err := c.RequestWithContext(ctx, http.MethodPatch, resource, requestBody)
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

// PutWithContext performs a PUT request with context support.
func (c *DataServiceClient) PutWithContext(ctx context.Context, resource string, requestBody string, responseBodyOut any) (*http.Response, error) {
	request, err := c.RequestWithContext(ctx, http.MethodPut, resource, requestBody)
	if err != nil {
		return nil, err
	}
	return c.Do(request, responseBodyOut)
}

// DeleteWithContext performs a DELETE request with context support.
func (c *DataServiceClient) DeleteWithContext(ctx context.Context, resource string) (*http.Response, error) {
	request, err := c.RequestWithContext(ctx, http.MethodDelete, resource, "")
	if err != nil {
		return nil, err
	}
	return c.Do(request, nil)
}

// PostJSONWithContext marshals body to JSON and performs POST with context.
func (c *DataServiceClient) PostJSONWithContext(ctx context.Context, resource string, body any, responseBodyOut any) (*http.Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.PostWithContext(ctx, resource, string(bodyBytes), responseBodyOut)
}

// PatchJSONWithContext marshals body to JSON and performs PATCH with context.
func (c *DataServiceClient) PatchJSONWithContext(ctx context.Context, resource string, body any, responseBodyOut any) (*http.Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.PatchWithContext(ctx, resource, string(bodyBytes), responseBodyOut)
}

// PutJSONWithContext marshals body to JSON and performs PUT with context.
func (c *DataServiceClient) PutJSONWithContext(ctx context.Context, resource string, body any, responseBodyOut any) (*http.Response, error) {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	return c.PutWithContext(ctx, resource, string(bodyBytes), responseBodyOut)
}
