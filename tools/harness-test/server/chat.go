package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (s *MockServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)

	var req llmRequest
	json.Unmarshal(body, &req)
	model := req.modelOrDefault()

	if s.shouldToolCall(req.hasTools()) {
		s.chatToolCall(w, model, req.Stream)
		return
	}

	text := s.getResponse()
	if req.Stream {
		s.chatStream(w, model, text)
		return
	}

	writeJSON(w, ChatCompletion{
		ID: "mock-1", Object: "chat.completion", Model: model,
		Choices: []ChatChoice{{
			Message:      &ChatMessage{Role: "assistant", Content: text},
			FinishReason: "stop",
		}},
		Usage: &Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	})
}

func (s *MockServer) chatStream(w http.ResponseWriter, model, text string) {
	f := beginSSE(w)

	streamData(w, f, ChatCompletion{
		ID: "mock-1", Object: "chat.completion.chunk", Model: model,
		Choices: []ChatChoice{{Delta: &ChatMessage{Role: "assistant", Content: text}}},
	})
	streamData(w, f, ChatCompletion{
		ID: "mock-1", Object: "chat.completion.chunk", Model: model,
		Choices: []ChatChoice{{Delta: &ChatMessage{}, FinishReason: "stop"}},
	})
	fmt.Fprint(w, "data: [DONE]\n\n")
	if f != nil {
		f.Flush()
	}
}

func (s *MockServer) chatToolCall(w http.ResponseWriter, model string, stream bool) {
	name, args := s.getToolCall()
	tc := ToolCall{
		ID:       "call_mock_1",
		Type:     "function",
		Function: FunctionCall{Name: name, Arguments: args},
	}

	if stream {
		f := beginSSE(w)
		streamData(w, f, ChatCompletion{
			ID: "mock-tc", Object: "chat.completion.chunk", Model: model,
			Choices: []ChatChoice{{Delta: &ChatMessage{Role: "assistant", ToolCalls: []ToolCall{tc}}}},
		})
		streamData(w, f, ChatCompletion{
			ID: "mock-tc", Object: "chat.completion.chunk", Model: model,
			Choices: []ChatChoice{{Delta: &ChatMessage{}, FinishReason: "tool_calls"}},
		})
		fmt.Fprint(w, "data: [DONE]\n\n")
		if f != nil {
			f.Flush()
		}
		return
	}

	writeJSON(w, ChatCompletion{
		ID: "mock-tc", Object: "chat.completion", Model: model,
		Choices: []ChatChoice{{
			Message:      &ChatMessage{Role: "assistant", ToolCalls: []ToolCall{tc}},
			FinishReason: "tool_calls",
		}},
		Usage: &Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
	})
}
