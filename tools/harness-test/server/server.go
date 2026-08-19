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
	Timestamp time.Time         `json:"ts"`
	Method    string            `json:"method"`
	Path      string            `json:"path"`
	Headers   map[string]string `json:"headers,omitempty"`
	Body      json.RawMessage   `json:"body,omitempty"`
	Model     string            `json:"model,omitempty"`
}

type MockServer struct {
	srv      *http.Server
	listener net.Listener

	mu           sync.Mutex
	log          []LogEntry
	response     string
	toolCallMode bool
	toolName     string
	toolArgs     string
}

func New() *MockServer {
	s := &MockServer{
		response: "Hello from mock server.",
	}
	mux := http.NewServeMux()

	// LLM APIs — each registered with and without /v1/ prefix
	for _, prefix := range []string{"/v1", ""} {
		mux.HandleFunc("POST "+prefix+"/chat/completions", s.handleChatCompletions)
		mux.HandleFunc("POST "+prefix+"/responses", s.handleResponses)
		mux.HandleFunc("POST "+prefix+"/messages", s.handleMessages)
		mux.HandleFunc("GET "+prefix+"/models", s.handleModels)
	}

	// Harness management endpoints (Grok auth, settings, sessions)
	stub := s.handleGrokJSON(map[string]any{})
	settings := s.handleGrokJSON(map[string]any{"models": map[string]any{"default": "mock-model"}})
	for _, prefix := range []string{"/v1", ""} {
		mux.HandleFunc("GET "+prefix+"/user", s.handleGrokJSON(map[string]any{"userId": "mock-user", "email": "mock@test.invalid"}))
		mux.HandleFunc("GET "+prefix+"/settings", settings)
		mux.HandleFunc("GET "+prefix+"/privacy/coding-data-retention", s.handleGrokJSON(map[string]any{"opted_out": false}))
	}
	mux.HandleFunc("GET /api-key", stub)
	mux.HandleFunc("GET /billing", stub)
	mux.HandleFunc("GET /feedback/config", stub)
	mux.HandleFunc("GET /bundle/archive", stub)
	mux.HandleFunc("POST /sessions/{id}/data", s.handleGrokRecord)
	mux.HandleFunc("PUT /sessions/{id}", s.handleGrokRecord)

	// Test utilities
	mux.HandleFunc("GET /log", s.handleGetLog)
	mux.HandleFunc("GET /log/count", s.handleLogCount)
	mux.HandleFunc("DELETE /log", s.handleClearLog)
	mux.HandleFunc("POST /response", s.handleSetResponse)

	// Catch-all: return 200 for any unhandled path (prevents 404 crashes)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.record(r, body)
		writeJSON(w, map[string]any{"ok": true})
	})

	s.srv = &http.Server{Handler: mux}
	return s
}

// --- Public API ---

func (s *MockServer) SetResponse(text string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.response = text
}

func (s *MockServer) SetToolCallMode(on bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolCallMode = on
}

func (s *MockServer) SetToolCall(name, args string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.toolName = name
	s.toolArgs = args
}

func (s *MockServer) LogCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.log)
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

// --- Internal helpers ---

func (s *MockServer) getResponse() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.response
}

func (s *MockServer) getToolCall() (string, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.toolName != "" {
		return s.toolName, s.toolArgs
	}
	return "Read", `{"file_path":"README.md"}`
}

func (s *MockServer) shouldToolCall(hasTools bool) bool {
	if !hasTools {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.toolCallMode {
		s.toolCallMode = false
		return true
	}
	return false
}

func (s *MockServer) record(r *http.Request, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	headers := make(map[string]string)
	for k, v := range r.Header {
		headers[strings.ToLower(k)] = strings.Join(v, ", ")
	}

	var model string
	var parsed struct {
		Model string `json:"model"`
	}
	if json.Unmarshal(body, &parsed) == nil {
		model = parsed.Model
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

// --- Handlers: models, test endpoints, Grok ---

func (s *MockServer) handleModels(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, ModelList{
		Data: []ModelEntry{
			{ID: "mock-model", Object: "model", OwnedBy: "mock"},
			{ID: "gpt-4o-mini", Object: "model", OwnedBy: "mock"},
		},
	})
}

func (s *MockServer) handleGetLog(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.Log())
}

func (s *MockServer) handleLogCount(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]int{"count": s.LogCount()})
}

func (s *MockServer) handleClearLog(w http.ResponseWriter, _ *http.Request) {
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

func (s *MockServer) handleGrokJSON(data any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.record(r, nil)
		writeJSON(w, data)
	}
}

func (s *MockServer) handleGrokRecord(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	s.record(r, body)
	w.WriteHeader(200)
	writeJSON(w, map[string]any{"ok": true})
}

// --- SSE helpers ---

type sseEvent struct {
	Type string
	Data any
}

func typed(eventType string, data map[string]any) sseEvent {
	data["type"] = eventType
	return sseEvent{Type: eventType, Data: data}
}

func beginSSE(w http.ResponseWriter) http.Flusher {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(200)
	f, _ := w.(http.Flusher)
	return f
}

func streamSSEEvents(w http.ResponseWriter, events []sseEvent) {
	f := beginSSE(w)
	for _, evt := range events {
		fmt.Fprintf(w, "event: %s\ndata: %s\n\n", evt.Type, mustJSON(evt.Data))
		if f != nil {
			f.Flush()
		}
	}
}

func streamRawSSE(w http.ResponseWriter, events []string) {
	f := beginSSE(w)
	for _, e := range events {
		fmt.Fprint(w, e)
		if f != nil {
			f.Flush()
		}
	}
}

func streamData(w http.ResponseWriter, f http.Flusher, v any) {
	fmt.Fprintf(w, "data: %s\n\n", mustJSON(v))
	if f != nil {
		f.Flush()
	}
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
