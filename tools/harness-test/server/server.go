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

	mu       sync.Mutex
	log      []LogEntry
	response string // default response text
}

func New() *MockServer {
	s := &MockServer{
		response: "Hello from mock server.",
	}
	mux := http.NewServeMux()

	// OpenAI Chat Completions
	mux.HandleFunc("POST /v1/chat/completions", s.handleChatCompletions)

	// OpenAI Responses API (WebSocket upgrade or HTTP fallback)
	mux.HandleFunc("POST /v1/responses", s.handleResponses)

	// Anthropic Messages
	mux.HandleFunc("POST /v1/messages", s.handleMessages)

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

// POST /v1/responses — OpenAI Responses API (HTTP fallback, not WebSocket)
func (s *MockServer) handleResponses(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)

	text := s.getResponse()

	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(200)

	fmt.Fprintf(w, "event: response.created\ndata: %s\n\n", mustJSON(map[string]any{
		"type": "response.created",
		"response": map[string]any{
			"id": "mock-resp-1", "status": "in_progress",
			"output": []any{},
		},
	}))
	fmt.Fprintf(w, "event: response.output_item.added\ndata: %s\n\n", mustJSON(map[string]any{
		"type":     "response.output_item.added",
		"output_index": 0,
		"item": map[string]any{
			"type": "message", "role": "assistant",
			"content": []any{map[string]any{"type": "output_text", "text": ""}},
		},
	}))
	fmt.Fprintf(w, "event: response.output_text.delta\ndata: %s\n\n", mustJSON(map[string]any{
		"type": "response.output_text.delta",
		"output_index": 0, "content_index": 0,
		"delta": text,
	}))
	fmt.Fprintf(w, "event: response.output_text.done\ndata: %s\n\n", mustJSON(map[string]any{
		"type": "response.output_text.done",
		"output_index": 0, "content_index": 0,
		"text": text,
	}))
	fmt.Fprintf(w, "event: response.completed\ndata: %s\n\n", mustJSON(map[string]any{
		"type": "response.completed",
		"response": map[string]any{
			"id": "mock-resp-1", "status": "completed",
			"output": []any{map[string]any{
				"type": "message", "role": "assistant",
				"content": []any{map[string]any{"type": "output_text", "text": text}},
			}},
			"usage": map[string]any{"input_tokens": 10, "output_tokens": 5, "total_tokens": 15},
		},
	}))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}
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

	if stream {
		w.Header().Set("Content-Type", "text/event-stream")
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
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
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
