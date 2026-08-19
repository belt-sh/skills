package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	Tools            []any                  `json:"tools,omitempty"`
	GenerationConfig map[string]any         `json:"generationConfig,omitempty"`
	SystemInstruction *geminiContent        `json:"systemInstruction,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text             string          `json:"text,omitempty"`
	FunctionCall     *geminiFuncCall `json:"functionCall,omitempty"`
	FunctionResponse *geminiFuncResp `json:"functionResponse,omitempty"`
}

type geminiFuncCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args"`
}

type geminiFuncResp struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata *geminiUsage      `json:"usageMetadata,omitempty"`
	ModelVersion  string            `json:"modelVersion,omitempty"`
}

type geminiCandidate struct {
	Content      geminiContent `json:"content"`
	FinishReason string        `json:"finishReason,omitempty"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

func (s *MockServer) handleGemini(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req geminiRequest
	json.Unmarshal(body, &req)

	model := "gemini-2.5-flash"
	if parts := strings.Split(r.URL.Path, "/"); len(parts) >= 4 {
		model = strings.TrimSuffix(parts[3], ":streamGenerateContent")
		model = strings.TrimSuffix(model, ":generateContent")
	}
	s.record(r, body, model)

	hasTools := len(req.Tools) > 0
	if s.shouldToolCall(hasTools, r.URL.Path) {
		s.geminiToolCall(w, model)
		return
	}

	text := s.getResponse()
	if strings.Contains(r.URL.RawQuery, "alt=sse") {
		s.geminiStream(w, model, text)
		return
	}

	writeJSON(w, geminiResponse{
		Candidates: []geminiCandidate{{
			Content:      geminiContent{Role: "model", Parts: []geminiPart{{Text: text}}},
			FinishReason: "STOP",
		}},
		UsageMetadata: &geminiUsage{PromptTokenCount: 10, CandidatesTokenCount: 5, TotalTokenCount: 15},
		ModelVersion:  model,
	})
}

func (s *MockServer) geminiStream(w http.ResponseWriter, model, text string) {
	events := []sseEvent{
		{"message", geminiResponse{
			Candidates:   []geminiCandidate{{Content: geminiContent{Role: "model", Parts: []geminiPart{{Text: text}}}}},
			ModelVersion: model,
		}},
		{"message", geminiResponse{
			Candidates:    []geminiCandidate{{Content: geminiContent{Role: "model", Parts: []geminiPart{{Text: ""}}}, FinishReason: "STOP"}},
			UsageMetadata: &geminiUsage{PromptTokenCount: 10, CandidatesTokenCount: 5, TotalTokenCount: 15},
			ModelVersion:  model,
		}},
	}
	f := beginSSE(w)
	for _, evt := range events {
		fmt.Fprintf(w, "data: %s\n\n", mustJSON(evt.Data))
		if f != nil {
			f.Flush()
		}
	}
}

func (s *MockServer) geminiToolCall(w http.ResponseWriter, model string) {
	name, argsJSON := s.getToolCall()
	var args map[string]any
	json.Unmarshal([]byte(argsJSON), &args)

	fc := geminiPart{FunctionCall: &geminiFuncCall{Name: name, Args: args}}

	events := []sseEvent{
		{"message", geminiResponse{
			Candidates:   []geminiCandidate{{Content: geminiContent{Role: "model", Parts: []geminiPart{fc}}}},
			ModelVersion: model,
		}},
		{"message", geminiResponse{
			Candidates:    []geminiCandidate{{Content: geminiContent{Role: "model", Parts: []geminiPart{fc}}, FinishReason: "STOP"}},
			UsageMetadata: &geminiUsage{PromptTokenCount: 10, CandidatesTokenCount: 12, TotalTokenCount: 22},
			ModelVersion:  model,
		}},
	}
	f := beginSSE(w)
	for _, evt := range events {
		fmt.Fprintf(w, "data: %s\n\n", mustJSON(evt.Data))
		if f != nil {
			f.Flush()
		}
	}
}
