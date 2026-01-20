package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/rhl/businessos-backend/internal/database/sqlc"
	"github.com/rhl/businessos-backend/internal/tools"
)

// GroqToolCall represents a tool call in Groq's response
type GroqToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// GroqMessage represents a message in the conversation with tool support
type GroqMessage struct {
	Role       string         `json:"role"`
	Content    string         `json:"content,omitempty"`
	ToolCalls  []GroqToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	Name       string         `json:"name,omitempty"`
}

// GroqChoice represents a choice with tool support
type GroqChoice struct {
	Index   int         `json:"index"`
	Message GroqMessage `json:"message"`
}

// GroqResponseWithTools represents Groq API response with function calling
type GroqResponseWithTools struct {
	Choices []GroqChoice `json:"choices"`
}

// callGroqAPIWithTools calls Groq API with function calling support
func (h *Handlers) callGroqAPIWithTools(ctx context.Context, messages []VoiceChatMessage, userID string) (string, error) {
	groqBaseURL := "https://api.groq.com/openai/v1/chat/completions"
	groqAPIKey := h.cfg.GroqAPIKey

	if groqAPIKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY not configured")
	}

	// Convert messages to Groq format
	groqMessages := make([]GroqMessage, len(messages))
	for i, msg := range messages {
		groqMessages[i] = GroqMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Get tool definitions
	toolDefinitions := tools.GetAllTools()

	// Build request
	reqBody := map[string]interface{}{
		"model":       "llama-3.3-70b-versatile", // Groq recommended model for tool use (Jan 2026)
		"messages":    groqMessages,
		"tools":       toolDefinitions,
		"tool_choice": "auto",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	slog.Info("🤖 Calling Groq API with tools",
		"model", "llama-3.3-70b-versatile",
		"message_count", len(groqMessages),
		"tool_count", len(toolDefinitions),
	)

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", groqBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+groqAPIKey)

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Groq API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Groq API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var groqResp GroqResponseWithTools
	if err := json.Unmarshal(body, &groqResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	choice := groqResp.Choices[0]

	// Check if LLM wants to call tools
	if len(choice.Message.ToolCalls) > 0 {
		// Add limit check
		if len(choice.Message.ToolCalls) > 10 {
			return "", fmt.Errorf("too many tool calls requested: %d (max 10)", len(choice.Message.ToolCalls))
		}

		slog.Info("🔧 LLM requested tool calls",
			"count", len(choice.Message.ToolCalls),
		)

		// Execute all tool calls
		queries := sqlc.New(h.pool)
		toolExecutor := tools.NewToolExecutor(queries, slog.Default())
		toolResults := make([]GroqMessage, 0)

		// Add assistant message with tool calls
		groqMessages = append(groqMessages, choice.Message)

		for _, toolCall := range choice.Message.ToolCalls {
			slog.Info("⚙️ Executing tool",
				"tool", toolCall.Function.Name,
				"args", toolCall.Function.Arguments,
			)

			result, err := toolExecutor.ExecuteToolCall(ctx, tools.ToolCall{
				ID:   toolCall.ID,
				Type: toolCall.Type,
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      toolCall.Function.Name,
					Arguments: toolCall.Function.Arguments,
				},
			}, userID)

			if err != nil {
				slog.Error("Failed to execute tool", "tool", toolCall.Function.Name, "error", err)
				result = fmt.Sprintf("Error executing %s: %v", toolCall.Function.Name, err)
			}

			slog.Info("✅ Tool executed",
				"tool", toolCall.Function.Name,
				"result", result,
			)

			// Add tool result message
			toolResults = append(toolResults, GroqMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: toolCall.ID,
				Name:       toolCall.Function.Name,
			})
		}

		// Add tool results to messages
		groqMessages = append(groqMessages, toolResults...)

		// Call LLM again with tool results
		slog.Info("🔄 Calling LLM again with tool results",
			"tool_result_count", len(toolResults),
		)

		reqBody["messages"] = groqMessages
		reqBody["tools"] = toolDefinitions // Still provide tools for potential follow-up

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", fmt.Errorf("failed to marshal second request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", groqBaseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			return "", fmt.Errorf("failed to create second request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+groqAPIKey)

		resp, err = client.Do(req)
		if err != nil {
			return "", fmt.Errorf("failed to call Groq API (second call): %w", err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read second response: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("Groq API error on second call (status %d): %s", resp.StatusCode, string(body))
		}

		// DEBUG: Log raw response
		slog.Info("🔍 DEBUG: Raw Groq response after tool execution", "body", string(body)[:min(500, len(body))])

		if err := json.Unmarshal(body, &groqResp); err != nil {
			return "", fmt.Errorf("failed to parse second response: %w", err)
		}
		if len(groqResp.Choices) == 0 {
			return "", fmt.Errorf("no choices in second response")
		}

		finalResponse := groqResp.Choices[0].Message.Content

		// Strip out any XML function tags that Groq sometimes includes
		// Pattern: <function=navigate_to_module{"module":"tasks"}></function>
		functionTagRegex := regexp.MustCompile(`<function=\w+\{[^}]*\}></function>`)
		cleanResponse := functionTagRegex.ReplaceAllString(finalResponse, "")

		// Also strip any standalone XML-like tags
		cleanResponse = regexp.MustCompile(`</?function[^>]*>`).ReplaceAllString(cleanResponse, "")

		// Trim whitespace
		cleanResponse = strings.TrimSpace(cleanResponse)

		slog.Info("✅ Final response after tool execution",
			"response_length", len(cleanResponse),
			"response", cleanResponse,
		)

		return cleanResponse, nil
	}

	// No tool calls, return direct response
	return choice.Message.Content, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
