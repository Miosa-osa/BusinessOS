package osa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestChatWaitsForAsyncRuntimeAndDoesNotDuplicateFinalAnswer(t *testing.T) {
	dispatched := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/stream/conversation-1":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: connected\ndata: {}\n\n")
			w.(http.Flusher).Flush()
			<-dispatched
			fmt.Fprint(w, "event: streaming_token\ndata: {\"text\":\"Internal runtime text that must not be exposed\"}\n\nevent: agent_response\ndata: {\"response\":\"Hello\"}\n\nevent: done\ndata: {}\n\n")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/api/v1/orchestrate":
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"status":"processing","session_id":"conversation-1"}`)
			close(dispatched)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := NewClient(&Config{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	events, err := client.Chat(ctx, &OrchestrateRequest{SessionID: "conversation-1", Input: "Hello"})
	if err != nil {
		t.Fatal(err)
	}
	var answer string
	var done bool
	for event := range events {
		if event.Type == "streaming_token" {
			answer += event.Data["text"].(string)
		}
		if event.Type == EventResponse {
			answer += event.Data["content"].(string)
		}
		if event.Type == "done" {
			done = true
		}
		if event.Type == EventError {
			t.Fatalf("unexpected error: %v", event.Data)
		}
	}
	if answer != "Hello" || !done {
		t.Fatalf("answer=%q done=%v", answer, done)
	}
}

func TestRuntimeAuthAllowsOnlyLoopbackWithoutToken(t *testing.T) {
	for _, endpoint := range []string{"http://127.0.0.1:9089", "http://localhost:9089", "http://[::1]:9089"} {
		if err := (&Config{BaseURL: endpoint, Timeout: time.Second}).Validate(); err != nil {
			t.Errorf("%s: %v", endpoint, err)
		}
	}
	for _, endpoint := range []string{"https://osa.example.com", "http://192.168.1.20:9089", "http://localhost.example.com"} {
		if err := (&Config{BaseURL: endpoint, Timeout: time.Second}).Validate(); err == nil {
			t.Errorf("accepted unauthenticated remote %s", endpoint)
		}
	}
}

func TestChatCancellationStopsRuntimeTurn(t *testing.T) {
	cancelled := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/stream/cancel-test":
			w.Header().Set("Content-Type", "text/event-stream")
			fmt.Fprint(w, "event: connected\ndata: {}\n\n")
			w.(http.Flusher).Flush()
			<-r.Context().Done()
		case "/api/v1/orchestrate":
			w.WriteHeader(http.StatusAccepted)
			fmt.Fprint(w, `{"status":"processing"}`)
		case "/api/v1/sessions/cancel-test/cancel":
			close(cancelled)
			fmt.Fprint(w, `{"status":"cancel_requested"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	client, err := NewClient(&Config{BaseURL: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	events, err := client.Chat(ctx, &OrchestrateRequest{SessionID: "cancel-test", Input: "test"})
	if err != nil {
		t.Fatal(err)
	}
	cancel()
	for range events {
	}
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("runtime cancellation was not requested")
	}
}

func TestChatConfirmsOverdriveBeforeDispatch(t *testing.T) {
	for _, confirmed := range []bool{true, false} {
		t.Run(fmt.Sprint(confirmed), func(t *testing.T) {
			modeSet, dispatched := false, false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/api/v1/commands/execute":
					var body map[string]string
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Error(err)
					}
					if body["session_id"] != "overdrive-test" || body["command"] != "permission_mode" || body["arg"] != "overdrive" {
						t.Errorf("wrong mode request: %v", body)
					}
					modeSet = true
					if confirmed {
						fmt.Fprint(w, `{"mode":"overdrive","output":"Permission mode set to overdrive."}`)
					} else {
						fmt.Fprint(w, `{"mode":"overdrive","output":"No active session"}`)
					}
				case "/api/v1/stream/overdrive-test":
					if !modeSet {
						t.Error("subscribed before configuring mode")
					}
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(w, "event: connected\ndata: {}\n\n")
					w.(http.Flusher).Flush()
					<-r.Context().Done()
				case "/api/v1/orchestrate":
					if !modeSet {
						t.Error("dispatched before configuring mode")
					}
					dispatched = true
					fmt.Fprint(w, `{"status":"completed","output":"Ready"}`)
				default:
					http.NotFound(w, r)
				}
			}))
			defer server.Close()
			client, err := NewClient(&Config{BaseURL: server.URL, Timeout: time.Second})
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			events, err := client.Chat(ctx, &OrchestrateRequest{SessionID: "overdrive-test", Input: "test", PermissionMode: "overdrive"})
			if confirmed {
				if err != nil {
					t.Fatal(err)
				}
				for range events {
				}
				if !dispatched {
					t.Fatal("confirmed request was not dispatched")
				}
			} else {
				if err == nil || dispatched {
					t.Fatal("unconfirmed overdrive was dispatched")
				}
			}
		})
	}
}
