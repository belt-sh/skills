package server

import (
	"encoding/json"
	"io"
	"net/http"
)

func (s *MockServer) handleResponses(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)

	var req llmRequest
	json.Unmarshal(body, &req)

	if s.shouldToolCall(req.hasTools()) {
		s.responsesToolCall(w)
		return
	}

	s.responsesText(w, s.getResponse())
}

func (s *MockServer) responsesText(w http.ResponseWriter, text string) {
	part := ResponsePart{Type: "output_text", Text: text}
	msg := ResponseItem{Type: "message", Role: "assistant", Content: []ResponsePart{part}}
	emptyPart := ResponsePart{Type: "output_text"}
	emptyMsg := ResponseItem{Type: "message", Role: "assistant", Content: []ResponsePart{emptyPart}}

	events := []sseEvent{
		typed("response.created", map[string]any{
			"response": ResponseObject{ID: "mock-resp-1", Status: "in_progress"},
		}),
		typed("response.output_item.added", map[string]any{
			"output_index": 0, "item": emptyMsg,
		}),
		typed("response.content_part.added", map[string]any{
			"output_index": 0, "content_index": 0, "part": emptyPart,
		}),
		typed("response.output_text.delta", map[string]any{
			"output_index": 0, "content_index": 0, "delta": text,
		}),
		typed("response.output_text.done", map[string]any{
			"output_index": 0, "content_index": 0, "text": text,
		}),
		typed("response.content_part.done", map[string]any{
			"output_index": 0, "content_index": 0, "part": part,
		}),
		typed("response.output_item.done", map[string]any{
			"output_index": 0, "item": msg,
		}),
		typed("response.completed", map[string]any{
			"response": ResponseObject{
				ID: "mock-resp-1", Status: "completed",
				Output: []ResponseItem{msg},
				Usage:  &Usage{InputTokens: 10, OutputTokens: 5, TotalTokens: 15},
			},
		}),
	}
	streamSSEEvents(w, events)
}

func (s *MockServer) responsesToolCall(w http.ResponseWriter) {
	name, args := s.getToolCall()
	fc := ResponseItem{
		Type: "function_call", ID: "fc_mock_1", CallID: "call_mock_1",
		Name: name, Status: "in_progress",
	}
	fcDone := ResponseItem{
		Type: "function_call", ID: "fc_mock_1", CallID: "call_mock_1",
		Name: name, Arguments: args, Status: "completed",
	}

	events := []sseEvent{
		typed("response.created", map[string]any{
			"response": ResponseObject{ID: "mock-resp-tc", Status: "in_progress"},
		}),
		typed("response.output_item.added", map[string]any{
			"response_id": "mock-resp-tc", "output_index": 0, "item": fc,
		}),
		typed("response.function_call_arguments.delta", map[string]any{
			"response_id": "mock-resp-tc", "item_id": "fc_mock_1", "output_index": 0, "delta": args,
		}),
		typed("response.function_call_arguments.done", map[string]any{
			"response_id": "mock-resp-tc", "item_id": "fc_mock_1", "output_index": 0, "arguments": args,
		}),
		typed("response.output_item.done", map[string]any{
			"response_id": "mock-resp-tc", "output_index": 0, "item": fcDone,
		}),
		typed("response.completed", map[string]any{
			"response": ResponseObject{
				ID: "mock-resp-tc", Status: "completed",
				Output: []ResponseItem{fcDone},
				Usage:  &Usage{InputTokens: 10, OutputTokens: 12, TotalTokens: 22},
			},
		}),
	}
	streamSSEEvents(w, events)
}
