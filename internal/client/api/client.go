package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"goph-keeper/internal/contract"
)

// Client вызывает HTTP API GophKeeper.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient создаёт API-клиент.
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ErrConflict — сервер вернул 409 с более новыми версиями.
var ErrConflict = errors.New("api: sync conflict")

// Error — HTTP-ошибка API.
type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("api: http %d: %s", e.Status, e.Message)
}

// Is сообщает errors.Is, что 409 — это ErrConflict, в том числе через %w.
func (e *Error) Is(target error) bool {
	return target == ErrConflict && e.Status == http.StatusConflict
}

func (client *Client) postJSON(ctx context.Context, path string, body any, accessToken string, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	response, err := client.do(request)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	return decodeResponse(response, out)
}

func (client *Client) do(req *http.Request) (*http.Response, error) {
	httpClient := client.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return httpClient.Do(req)
}

func decodeResponse(resp *http.Response, out any) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errBody contract.ErrorResponse
		if len(body) > 0 {
			_ = json.Unmarshal(body, &errBody)
		}
		return &Error{Status: resp.StatusCode, Message: errBody.Error}
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("api: decode response: %w", err)
	}
	return nil
}

func readAPIError(resp *http.Response) error {
	var out contract.ErrorResponse
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &out)
	}
	return &Error{Status: resp.StatusCode, Message: out.Error}
}
