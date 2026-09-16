package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var ErrNoToken = errors.New("no token: set FOLIO_TOKEN or run `folio login`")

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

type apiError struct {
	status  int
	Message string `json:"message"`
	Data    map[string]struct {
		Message string `json:"message"`
	} `json:"data"`
}

func (e apiError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("%d: request failed", e.status)
	}
	return fmt.Sprintf("%d: %s", e.status, e.Message)
}

func (c *Client) do(method, path string, body any, out any) error {
	if c.token == "" {
		return ErrNoToken
	}

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		payload = bytes.NewReader(encoded)
	}

	req, err := http.NewRequest(method, c.baseURL+path, payload)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode >= 400 {
		failure := apiError{status: res.StatusCode}
		_ = json.NewDecoder(res.Body).Decode(&failure)
		return failure
	}

	if out == nil || res.StatusCode == http.StatusNoContent {
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}
