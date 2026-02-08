package discord

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookClient_Notify(t *testing.T) {
	var receivedPayload map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedPayload)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := &WebhookClient{
		httpClient: server.Client(),
		webhookURL: server.URL,
	}

	err := client.Notify(context.Background(), "Test notification message")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedPayload["content"] != "Test notification message" {
		t.Errorf("expected 'Test notification message', got '%s'", receivedPayload["content"])
	}
}

func TestWebhookClient_Notify_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &WebhookClient{
		httpClient: server.Client(),
		webhookURL: server.URL,
	}

	err := client.Notify(context.Background(), "Test message")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
