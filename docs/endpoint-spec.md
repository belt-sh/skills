# Test Endpoint Specification

A minimal LLM API server that accepts requests from any harness, returns valid responses, and logs everything for test assertions.

## Routes

### `POST /v1/chat/completions` (OpenAI format)

Request:
```json
{
  "model": "test-model",
  "messages": [{"role": "user", "content": "hello"}],
  "stream": true
}
```

Response (streaming SSE):
```
data: {"id":"test-1","object":"chat.completion.chunk","choices":[{"delta":{"role":"assistant","content":"Hello"},"index":0}]}

data: {"id":"test-1","object":"chat.completion.chunk","choices":[{"delta":{"content":" from test endpoint."},"index":0}]}

data: {"id":"test-1","object":"chat.completion.chunk","choices":[{"delta":{},"finish_reason":"stop","index":0}]}

data: [DONE]
```

Non-streaming response:
```json
{
  "id": "test-1",
  "object": "chat.completion",
  "model": "test-model",
  "choices": [{
    "message": {"role": "assistant", "content": "Hello from test endpoint."},
    "finish_reason": "stop",
    "index": 0
  }],
  "usage": {"prompt_tokens": 10, "completion_tokens": 5, "total_tokens": 15}
}
```

### `POST /v1/messages` (Anthropic format)

Request:
```json
{
  "model": "test-model",
  "messages": [{"role": "user", "content": "hello"}],
  "max_tokens": 1024
}
```

Response (streaming SSE):
```
event: message_start
data: {"type":"message_start","message":{"id":"test-1","type":"message","role":"assistant","content":[],"model":"test-model"}}

event: content_block_start
data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}

event: content_block_delta
data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello from test endpoint."}}

event: content_block_stop
data: {"type":"content_block_stop","index":0}

event: message_delta
data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":5}}

event: message_stop
data: {"type":"message_stop"}
```

### `POST /v1/models/*/generateContent` (Google format)

Minimal Gemini-compatible response. TBD — needs research per harness.

### `GET /v1/models`

Returns a minimal model list so harnesses that enumerate models don't error:
```json
{
  "data": [{"id": "test-model", "object": "model", "owned_by": "test"}]
}
```

## Logging

Every request is appended to a JSONL file:

```json
{"ts":"2026-08-16T00:00:00Z","method":"POST","path":"/v1/chat/completions","headers":{"authorization":"Bearer test-key","content-type":"application/json"},"body":{...},"remote_addr":"172.17.0.2"}
```

## Assertion helpers

The test endpoint also serves:

- `GET /log` — returns the full JSONL log
- `GET /log/count` — returns `{"count": N}`
- `DELETE /log` — clears the log (for test isolation)

## Configuration

```
TEST_ENDPOINT_PORT=4100       # listen port
TEST_ENDPOINT_LOG=/tmp/test-endpoint.jsonl
TEST_ENDPOINT_API_KEY=test-key  # accept any key, or require this one
```

## Implementation notes

- Single Go binary, no dependencies
- ~200 lines
- Can live in `go/cli/cmd/testendpoint/` or a standalone `tools/test-endpoint/`
