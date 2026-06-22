package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"goph-keeper/internal/contract"
)

// Pull загружает записи с version > since.
func (client *Client) Pull(ctx context.Context, accessToken string, since int64) ([]contract.Record, error) {
	url, err := url.Parse(client.BaseURL + "/api/v1/sync/")
	if err != nil {
		return nil, err
	}
	query := url.Query()
	query.Set("since", strconv.FormatInt(since, 10))
	url.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := client.do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	var out contract.PullResponse
	if err := decodeResponse(response, &out); err != nil {
		return nil, fmt.Errorf("api: pull: %w", err)
	}
	if out.Records == nil {
		return []contract.Record{}, nil
	}
	return out.Records, nil
}

// Push отправляет записи на сервер; при конфликте возвращает conflicts и ErrConflict.
func (client *Client) Push(ctx context.Context, accessToken string, records []contract.Record) ([]contract.Record, error) {
	body, err := json.Marshal(contract.PushRequest{Records: records})
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, client.BaseURL+"/api/v1/sync/", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+accessToken)

	response, err := client.do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()

	var out contract.PushResponse
	if err := decodeResponse(response, &out); err != nil {
		if errors.Is(err, ErrConflict) {
			if out.Conflicts == nil {
				out.Conflicts = []contract.Record{}
			}
			return out.Conflicts, ErrConflict
		}
		return nil, fmt.Errorf("api: push: %w", err)
	}
	return nil, nil
}
