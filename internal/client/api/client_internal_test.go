package api

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func TestDecodeJSONAndReadAPIError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       io.NopCloser(bytes.NewBufferString(`{"error":"bad request"}`)),
	}
	err := readAPIError(resp)
	if err == nil || err.Error() == "" {
		t.Fatal("expected api error")
	}
}

func TestDecodeJSONHandlesEmptyBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(bytes.NewBuffer(nil))}
	var out map[string]any
	if err := decodeResponse(resp, &out); err != nil {
		t.Fatal(err)
	}
}

func TestDecodeResponseReturnsErrorOnMalformedSuccessBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{not-json`)),
	}
	var out map[string]any
	if err := decodeResponse(resp, &out); err == nil {
		t.Fatal("expected decode error for malformed success body")
	}
}

func TestDecodeResponseDecodesSuccessBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewBufferString(`{"id":"abc"}`)),
	}
	var out struct {
		ID string `json:"id"`
	}
	if err := decodeResponse(resp, &out); err != nil {
		t.Fatal(err)
	}
	if out.ID != "abc" {
		t.Fatalf("expected decoded id, got %q", out.ID)
	}
}
