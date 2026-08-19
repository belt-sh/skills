package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *MockServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)

	var req llmRequest
	json.Unmarshal(body, &req)
	model := req.modelOrDefault()

	if s.shouldToolCall(req.hasTools(), r.URL.Path) {
		s.anthropicToolCall(w, model, req.Stream)
		return
	}

	text := s.getResponse()
	if req.Stream {
		streamRawSSE(w, anthropicTextEvents(model, text))
		return
	}

	writeJSON(w, AnthropicMessage{
		ID: "mock-1", Type: "message", Role: "assistant", Model: model,
		Content:    []ContentBlock{{Type: "text", Text: text}},
		StopReason: "end_turn",
		Usage:      &AntUsage{InputTokens: 10, OutputTokens: 5},
	})
}

func (s *MockServer) anthropicToolCall(w http.ResponseWriter, model string, stream bool) {
	name, args := s.getToolCall()
	var input any
	json.Unmarshal([]byte(args), &input)

	block := ContentBlock{Type: "tool_use", ID: "toolu_mock_1", Name: name, Input: input}

	if stream {
		streamRawSSE(w, []string{
			antEvent("message_start", map[string]any{
				"type": "message_start",
				"message": AnthropicMessage{
					ID: "mock-tc", Type: "message", Role: "assistant", Model: model,
				},
			}),
			antEvent("content_block_start", map[string]any{
				"type": "content_block_start", "index": 0,
				"content_block": ContentBlock{Type: "tool_use", ID: "toolu_mock_1", Name: name},
			}),
			antEvent("content_block_delta", map[string]any{
				"type": "content_block_delta", "index": 0,
				"delta": map[string]string{"type": "input_json_delta", "partial_json": args},
			}),
			antEvent("content_block_stop", map[string]any{
				"type": "content_block_stop", "index": 0,
			}),
			antEvent("message_delta", map[string]any{
				"type":  "message_delta",
				"delta": map[string]string{"stop_reason": "tool_use"},
				"usage": AntUsage{OutputTokens: 15},
			}),
			"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
		})
		return
	}

	writeJSON(w, AnthropicMessage{
		ID: "mock-tc", Type: "message", Role: "assistant", Model: model,
		Content:    []ContentBlock{block},
		StopReason: "tool_use",
		Usage:      &AntUsage{InputTokens: 10, OutputTokens: 15},
	})
}

func anthropicTextEvents(model, text string) []string {
	return []string{
		antEvent("message_start", map[string]any{
			"type": "message_start",
			"message": AnthropicMessage{
				ID: "mock-1", Type: "message", Role: "assistant", Model: model,
			},
		}),
		antEvent("content_block_start", map[string]any{
			"type": "content_block_start", "index": 0,
			"content_block": ContentBlock{Type: "text"},
		}),
		antEvent("content_block_delta", map[string]any{
			"type": "content_block_delta", "index": 0,
			"delta": map[string]string{"type": "text_delta", "text": text},
		}),
		antEvent("content_block_stop", map[string]any{
			"type": "content_block_stop", "index": 0,
		}),
		antEvent("message_delta", map[string]any{
			"type":  "message_delta",
			"delta": map[string]string{"stop_reason": "end_turn"},
			"usage": AntUsage{OutputTokens: 5},
		}),
		"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
	}
}

func antEvent(event string, data any) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event, mustJSON(data))
}
