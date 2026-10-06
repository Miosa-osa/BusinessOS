package osa

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Chat subscribes before dispatch and waits for OSA's terminal event, not its
// HTTP 202 acknowledgement. A turn is dispatched once: retries could run tools
// twice. The session ID is supplied by the authenticated BusinessOS conversation.
func (c *Client) Chat(ctx context.Context, req *OrchestrateRequest) (<-chan Event, error) {
	// Set the same session policy as `osa overdrive` before any tool can run.
	// OSA retains this setting even when the session loop is not created yet.
	if req.PermissionMode != "" {
		var mode struct {
			Mode   string `json:"mode"`
			Output string `json:"output"`
		}
		err := c.doJSON(ctx, http.MethodPost, "/api/v1/commands/execute", map[string]string{
			"command": "permission_mode", "arg": req.PermissionMode, "session_id": req.SessionID,
		}, &mode)
		if err != nil {
			return nil, fmt.Errorf("set OSA permission mode: %w", err)
		}
		if mode.Mode != req.PermissionMode || mode.Output != "Permission mode set to "+req.PermissionMode+"." {
			return nil, fmt.Errorf("OSA did not confirm %s mode: %s", req.PermissionMode, mode.Output)
		}
	}
	streamCtx, cancel := context.WithCancel(ctx)
	events, err := c.Stream(streamCtx, req.SessionID)
	if err != nil {
		cancel()
		return nil, err
	}
	response, err := c.Orchestrate(streamCtx, req)
	if err != nil {
		cancel()
		return nil, err
	}
	out := make(chan Event, 64)
	go func() {
		terminal := false
		defer close(out)
		defer cancel()
		defer func() {
			if !terminal && ctx.Err() != nil {
				stopCtx, stop := context.WithTimeout(context.Background(), 3*time.Second)
				defer stop()
				_ = c.doJSON(stopCtx, http.MethodPost, "/api/v1/sessions/"+url.PathEscape(req.SessionID)+"/cancel", nil, nil)
			}
		}()
		send := func(e Event) bool {
			select {
			case out <- e:
				return true
			case <-ctx.Done():
				return false
			}
		}
		if response.Status != "processing" && response.Status != "queued" && response.Status != "running" {
			text := response.Output
			if text == "" {
				text = response.Prompt
			}
			if text != "" {
				send(Event{Type: EventResponse, Data: map[string]interface{}{"content": text}})
			} else {
				send(Event{Type: EventError, Data: map[string]interface{}{"message": fmt.Sprintf("OSA did not start a turn (status %q)", response.Status)}})
			}
			terminal = true
			send(Event{Type: "done", Data: map[string]interface{}{}})
			return
		}
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-events:
				if !ok {
					return
				}
				switch e.Type {
				case "streaming_token":
					// This runtime does not distinguish reasoning from answer tokens.
					// Its terminal agent_response is the authoritative public answer.
					continue
				case "agent_response":
					if turnError := e.Data["turn_error"]; turnError != nil {
						send(Event{Type: EventError, Data: map[string]interface{}{"message": fmt.Sprint(turnError)}})
					} else {
						text, _ := e.Data["response"].(string)
						if text != "" {
							send(Event{Type: EventResponse, Data: map[string]interface{}{"content": text}})
						}
					}
					continue
				}
				if e.Type == "done" {
					terminal = true
				}
				if !send(e) {
					return
				}
				if e.Type == "done" || e.Type == EventError {
					return
				}
			}
		}
	}()
	return out, nil
}

// Chat deliberately does not use the retry/fallback orchestration path.
func (r *ResilientClient) Chat(ctx context.Context, req *OrchestrateRequest) (<-chan Event, error) {
	return r.client.Chat(ctx, req)
}

func (r *ResilientClient) UpdateModel(ctx context.Context, provider, model string) (*HealthResponse, error) {
	var result HealthResponse
	err := r.client.doJSON(ctx, http.MethodPost, "/api/v1/models/switch", map[string]string{"provider": provider, "model": model}, &result)
	if err != nil {
		return nil, err
	}
	if result.Provider != provider || result.Model != model {
		return nil, fmt.Errorf("OSA did not confirm the requested model")
	}
	r.InvalidateHealthCache()
	return &result, nil
}
