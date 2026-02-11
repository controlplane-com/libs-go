package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// get performs a GET request and unmarshals the response into result.
func (c *Client) get(ctx context.Context, path string, result any) error {
	return c.do(ctx, http.MethodGet, path, nil, result)
}

// post performs a POST request with the given body and unmarshals the response into result.
func (c *Client) post(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPost, path, body, result)
}

// patch performs a PATCH request with the given body and unmarshals the response into result.
func (c *Client) patch(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPatch, path, body, result)
}

// put performs a PUT request with the given body and unmarshals the response into result.
func (c *Client) put(ctx context.Context, path string, body, result any) error {
	return c.do(ctx, http.MethodPut, path, body, result)
}

// delete performs a DELETE request.
func (c *Client) delete(ctx context.Context, path string) error {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

// do performs an HTTP request with the given method, path, body, and result.
func (c *Client) do(ctx context.Context, method, path string, body, result any) error {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	c.setRequestHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return c.parseErrorResponse(resp.StatusCode, respBody)
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

// setRequestHeaders sets the standard headers for all requests.
func (c *Client) setRequestHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	} else {
		req.Header.Set("User-Agent", "cpln-go-client/1.0")
	}

	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
}

// parseErrorResponse parses an error response from the API.
func (c *Client) parseErrorResponse(statusCode int, body []byte) error {
	apiErr := &APIError{
		Status: statusCode,
	}

	if len(body) > 0 {
		if err := json.Unmarshal(body, apiErr); err != nil {
			// If we can't parse the error, use the body as the message
			apiErr.Message = strings.TrimSpace(string(body))
		}
	}

	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(statusCode)
	}

	return apiErr
}

// doRaw performs an HTTP request and returns the raw response body.
func (c *Client) doRaw(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setRequestHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.parseErrorResponse(resp.StatusCode, respBody)
	}

	return respBody, nil
}
