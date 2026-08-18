package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type LogEntry struct {
	Timestamp     time.Time         `json:"ts"`
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          json.RawMessage   `json:"body,omitempty"`
	Model         string            `json:"model,omitempty"`
	ResponseModel string           `json:"response_model,omitempty"`
}

type MockServer struct {
	srv      *http.Server
	listener net.Listener

	mu           sync.Mutex
	log          []LogEntry
	response     string // default response text
	toolCallMode bool   // respond with a tool call first, then text
	toolCallSeen bool   // track if we already sent the tool call
}

func New() *MockServer {
	s := &MockServer{
		response: "Hello from mock server.",
	}
	mux := http.NewServeMux()

	// OpenAI Chat Completions
	mux.HandleFunc("POST /v1/chat/completions", s.handleChatCompletions)

	// OpenAI Responses API
	mux.HandleFunc("POST /v1/responses", s.handleResponses)
	mux.HandleFunc("POST /responses", s.handleResponses)

	// Anthropic Messages
	mux.HandleFunc("POST /v1/messages", s.handleMessages)
	mux.HandleFunc("POST /messages", s.handleMessages)

	// Model listing
	mux.HandleFunc("GET /v1/models", s.handleModels)

	// Grok-specific endpoints
	mux.HandleFunc("GET /v1/user", s.handleUser)
	mux.HandleFunc("GET /v1/settings", s.handleSettings)
	mux.HandleFunc("GET /v1/privacy/coding-data-retention", s.handlePrivacy)
	mux.HandleFunc("POST /sessions/{id}/data", s.handleSessionData)
	mux.HandleFunc("PUT /sessions/{id}", s.handleSessionUpdate)

	// Test utilities
	mux.HandleFunc("GET /log", s.handleGetLog)
	mux.HandleFunc("GET /log/count", s.handleLogCount)
	mux.HandleFunc("DELETE /log", s.handleClearLog)
	mux.HandleFunc("POST /response", s.handleSetResponse)

	s.srv = &http.Server{Handler: mux}
	return s
}

func (s *MockServer) SetResponse(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.response = text
}

func (s *MockServer) SetToolCallMode(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolCallMode = on
	s.toolCallSeen = false
}

func (s *MockServer) shouldToolCall() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.toolCallMode && !s.toolCallSeen {
		s.toolCallSeen = true
		return true
	}
	return false
}

func (s *MockServer) Log() []LogEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]LogEntry, len(s.log))
	copy(cp, s.log)
	return cp
}

func (s *MockServer) ClearLog() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log = nil
}

func (s *MockServer) Start() (string, error) {
	var err error
	s.listener, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	go s.srv.Serve(s.listener)
	return fmt.Sprintf("http://%s", s.listener.Addr()), nil
}

func (s *MockServer) Close() {
	if s.srv != nil {
		s.srv.Close()
	}
}

func (s *MockServer) record(r *http.Request, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[strings.ToLower(k)] = strings.Join(v, ", ")
	}

	var model string
	var parsed map[string]any
	if json.Unmarshal(body, &parsed) == nil {
		if m, ok := parsed["model"].(string); ok {
			model = m
		}
	}

	s.log = append(s.log, LogEntry{
		Timestamp: time.Now(),
		Method:    r.Method,
		Path:      r.URL.Path,
		Headers:   headers,
		Body:      json.RawMessage(body),
		Model:     model,
	})
}

func (s *MockServer) getResponse() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.response
}

// POST /v1/chat/completions — OpenAI format
func (s *MockServer) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)

	var req map[string]any
	json.Unmarshal(body, &req)

	stream, _ := req["stream"].(bool)
	model, _ := req["model"].(string)
	if model == "" {
		model = "mock-model"
	}

	text := s.getResponse()

	if s.shouldToolCall() {
		s.handleChatToolCall(w, model, stream)
		return
	}

	if stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)

		fmt.Fprintf(w, "data: %s\n\n", mustJSON(map[string]any{
			"id": "mock-1", "object": "chat.completion.chunk", "model": model,
			"choices": []any{map[string]any{
				"index": 0, "delta": map[string]any{"role": "assistant", "content": text},
			}},
		}))
		fmt.Fprintf(w, "data: %s\n\n", mustJSON(map[string]any{
			"id": "mock-1", "object": "chat.completion.chunk", "model": model,
			"choices": []any{map[string]any{
				"index": 0, "delta": map[string]any{}, "finish_reason": "stop",
			}},
		}))
		fmt.Fprint(w, "data: [DONE]\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// stream done — flush is sufficient, [DONE] / message_stop signals EOF
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id": "mock-1", "object": "chat.completion", "model": model,
		"choices": []any{map[string]any{
			"message":       map[string]any{"role": "assistant", "content": text},
			"finish_reason": "stop", "index": 0,
		}},
		"usage": map[string]any{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15},
	})
}

func (s *MockServer) handleChatToolCall(w http.ResponseWriter, model string, stream bool) {
	toolCall := map[string]any{
		"id":   "call_mock_1",
		"type": "function",
		"function": map[string]any{
			"name":      "Read",
			"arguments": `{"file_path":"README.md"}`,
		},
	}
	if stream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)
		fmt.Fprintf(w, "data: %s\n\n", mustJSON(map[string]any{
			"id": "mock-tc", "object": "chat.completion.chunk", "model": model,
			"choices": []any{map[string]any{
				"index": 0, "delta": map[string]any{"role": "assistant", "tool_calls": []any{toolCall}},
			}},
		}))
		fmt.Fprintf(w, "data: %s\n\n", mustJSON(map[string]any{
			"id": "mock-tc", "object": "chat.completion.chunk", "model": model,
			"choices": []any{map[string]any{
				"index": 0, "delta": map[string]any{}, "finish_reason": "tool_calls",
			}},
		}))
		fmt.Fprint(w, "data: [DONE]\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// stream done — flush is sufficient, [DONE] / message_stop signals EOF
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id": "mock-tc", "object": "chat.completion", "model": model,
		"choices": []any{map[string]any{
			"message":       map[string]any{"role": "assistant", "tool_calls": []any{toolCall}},
			"finish_reason": "tool_calls", "index": 0,
		}},
		"usage": map[string]any{"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15},
	})
}

// POST /v1/responses and /responses — OpenAI Responses API
func (s *MockServer) handleResponses(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)

	text := s.getResponse()
	flusher, canFlush := w.(http.Flusher)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(200)

	events := []map[string]any{
		{"type": "response.created", "response": map[string]any{"id": "mock-resp-1", "status": "in_progress", "output": []any{}}},
		{"type": "response.output_item.added", "output_index": 0, "item": map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": ""}}}},
		{"type": "response.content_part.added", "output_index": 0, "content_index": 0, "part": map[string]any{"type": "output_text", "text": ""}},
		{"type": "response.output_text.delta", "output_index": 0, "content_index": 0, "delta": text},
		{"type": "response.output_text.done", "output_index": 0, "content_index": 0, "text": text},
		{"type": "response.content_part.done", "output_index": 0, "content_index": 0, "part": map[string]any{"type": "output_text", "text": text}},
		{"type": "response.output_item.done", "output_index": 0, "item": map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": text}}}},
		{"type": "response.completed", "response": map[string]any{"id": "mock-resp-1", "status": "completed", "output": []any{map[string]any{"type": "message", "role": "assistant", "content": []any{map[string]any{"type": "output_text", "text": text}}}}, "usage": map[string]any{"input_tokens": 10, "output_tokens": 5, "total_tokens": 15}}},
	}

	for _, evt := range events {
		evtType, _ := evt["type"].(string)
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evtType, mustJSON(evt))
		if canFlush {
			flusher.Flush()
		}
	}
	// response.completed signals EOF — client closes the reader
}

// POST /v1/messages — Anthropic format
func (s *MockServer) handleMessages(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)

	var req map[string]any
	json.Unmarshal(body, &req)

	stream := false
	if s, ok := req["stream"].(bool); ok {
		stream = s
	}
	model, _ := req["model"].(string)
	if model == "" {
		model = "mock-model"
	}

	text := s.getResponse()

	if s.shouldToolCall() {
		s.handleAnthropicToolCall(w, model, stream)
		return
	}

	if stream {
		flusher, canFlush := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(200)

		events := []string{
			fmt.Sprintf("event: message_start\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "message_start",
				"message": map[string]any{"id": "mock-1", "type": "message", "role": "assistant", "content": []any{}, "model": model},
			})),
			fmt.Sprintf("event: content_block_start\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "content_block_start", "index": 0,
				"content_block": map[string]any{"type": "text", "text": ""},
			})),
			fmt.Sprintf("event: content_block_delta\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "content_block_delta", "index": 0,
				"delta": map[string]any{"type": "text_delta", "text": text},
			})),
			fmt.Sprintf("event: content_block_stop\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "content_block_stop", "index": 0,
			})),
			fmt.Sprintf("event: message_delta\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "message_delta",
				"delta": map[string]any{"stop_reason": "end_turn"},
				"usage": map[string]any{"output_tokens": 5},
			})),
			"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
		}
		for _, e := range events {
			fmt.Fprint(w, e)
			if canFlush {
				flusher.Flush()
			}
		}
		// stream done — flush is sufficient, [DONE] / message_stop signals EOF
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id": "mock-1", "type": "message", "role": "assistant", "model": model,
		"content":     []any{map[string]any{"type": "text", "text": text}},
		"stop_reason": "end_turn",
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 5},
	})
}

func (s *MockServer) handleAnthropicToolCall(w http.ResponseWriter, model string, stream bool) {
	toolUseBlock := map[string]any{
		"type": "tool_use", "id": "toolu_mock_1", "name": "Read",
		"input": map[string]any{"file_path": "README.md"},
	}
	if stream {
		flusher, canFlush := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.WriteHeader(200)
		events := []string{
			fmt.Sprintf("event: message_start\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "message_start",
				"message": map[string]any{"id": "mock-tc", "type": "message", "role": "assistant", "content": []any{}, "model": model},
			})),
			fmt.Sprintf("event: content_block_start\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "content_block_start", "index": 0,
				"content_block": map[string]any{"type": "tool_use", "id": "toolu_mock_1", "name": "Read", "input": map[string]any{}},
			})),
			fmt.Sprintf("event: content_block_delta\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "content_block_delta", "index": 0,
				"delta": map[string]any{"type": "input_json_delta", "partial_json": `{"file_path":"README.md"}`},
			})),
			fmt.Sprintf("event: content_block_stop\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "content_block_stop", "index": 0,
			})),
			fmt.Sprintf("event: message_delta\ndata: %s\n\n", mustJSON(map[string]any{
				"type": "message_delta",
				"delta": map[string]any{"stop_reason": "tool_use"},
				"usage": map[string]any{"output_tokens": 15},
			})),
			"event: message_stop\ndata: {\"type\":\"message_stop\"}\n\n",
		}
		for _, e := range events {
			fmt.Fprint(w, e)
			if canFlush {
				flusher.Flush()
			}
		}
		// stream done — flush is sufficient, [DONE] / message_stop signals EOF
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id": "mock-tc", "type": "message", "role": "assistant", "model": model,
		"content":     []any{toolUseBlock},
		"stop_reason": "tool_use",
		"usage":       map[string]any{"input_tokens": 10, "output_tokens": 15},
	})
}

func (s *MockServer) handleModels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"data": []any{
			map[string]any{"id": "mock-model", "object": "model", "owned_by": "mock"},
			map[string]any{"id": "gpt-4o-mini", "object": "model", "owned_by": "mock"},
		},
	})
}

func (s *MockServer) handleGetLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.Log())
}

func (s *MockServer) handleLogCount(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"count": len(s.Log())})
}

func (s *MockServer) handleClearLog(w http.ResponseWriter, r *http.Request) {
	s.ClearLog()
	w.WriteHeader(204)
}

func (s *MockServer) handleSetResponse(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req struct {
		Text string `json:"text"`
	}
	if json.Unmarshal(body, &req) == nil && req.Text != "" {
		s.SetResponse(req.Text)
	}
	w.WriteHeader(204)
}

func (s *MockServer) handleUser(w http.ResponseWriter, r *http.Request) {
	s.record(r, nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"userId": "mock-user",
		"email":  "mock@test.invalid",
	})
}

func (s *MockServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	s.record(r, nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"models": map[string]any{
			"default": "mock-model",
		},
	})
}

func (s *MockServer) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	s.record(r, nil)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"opted_out": false,
	})
}

func (s *MockServer) handleSessionData(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func (s *MockServer) handleSessionUpdate(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]any{"ok": true})
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
