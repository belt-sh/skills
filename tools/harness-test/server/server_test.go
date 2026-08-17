package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestChatCompletions(t *testing.T) {
	s := New()
	s.SetResponse("test response")
	url, err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	body := `{"model":"test","messages":[{"role":"user","content":"hi"}]}`
	resp, err := http.Post(url+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}

	data, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(data, &result)

	choices, ok := result["choices"].([]any)
	if !ok || len(choices) == 0 {
		t.Fatal("no choices")
	}
	choice := choices[0].(map[string]any)
	msg := choice["message"].(map[string]any)
	if msg["content"] != "test response" {
		t.Fatalf("unexpected content: %v", msg["content"])
	}

	log := s.Log()
	if len(log) != 1 {
		t.Fatalf("expected 1 log entry, got %d", len(log))
	}
	if log[0].Model != "test" {
		t.Fatalf("logged model: %s", log[0].Model)
	}
}

func TestStreamingChatCompletions(t *testing.T) {
	s := New()
	s.SetResponse("streamed")
	url, err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	body := `{"model":"test","messages":[{"role":"user","content":"hi"}],"stream":true}`
	resp, err := http.Post(url+"/v1/chat/completions", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("Content-Type") != "text/event-stream" {
		t.Fatalf("wrong content type: %s", resp.Header.Get("Content-Type"))
	}

	data, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(data), "streamed") {
		t.Fatalf("response doesn't contain 'streamed': %s", data)
	}
	if !strings.Contains(string(data), "[DONE]") {
		t.Fatal("missing [DONE]")
	}
}

func TestAnthropicMessages(t *testing.T) {
	s := New()
	s.SetResponse("anthropic response")
	url, err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	body := `{"model":"claude","messages":[{"role":"user","content":"hi"}],"max_tokens":100}`
	resp, err := http.Post(url+"/v1/messages", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var result map[string]any
	json.Unmarshal(data, &result)

	content, ok := result["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatal("no content")
	}
	block := content[0].(map[string]any)
	if block["text"] != "anthropic response" {
		t.Fatalf("unexpected text: %v", block["text"])
	}
}

func TestResponsesAPI(t *testing.T) {
	s := New()
	s.SetResponse("responses api text")
	url, err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	body := `{"model":"test","input":"hi"}`
	resp, err := http.Post(url+"/v1/responses", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(data), "responses api text") {
		t.Fatalf("response doesn't contain expected text: %s", data)
	}
}

func TestLogEndpoints(t *testing.T) {
	s := New()
	url, err := s.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	// Make a request
	http.Post(url+"/v1/chat/completions", "application/json",
		strings.NewReader(`{"model":"test","messages":[{"role":"user","content":"hi"}]}`))

	// Check count
	resp, _ := http.Get(url + "/log/count")
	data, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(data), `"count":1`) {
		t.Fatalf("unexpected count: %s", data)
	}

	// Clear
	req, _ := http.NewRequest("DELETE", url+"/log", nil)
	http.DefaultClient.Do(req)

	// Verify cleared
	resp, _ = http.Get(url + "/log/count")
	data, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(data), `"count":0`) {
		t.Fatalf("log not cleared: %s", data)
	}
}
