package harness

var standardEvents = Events{
	SessionStart: "SessionStart",
	PromptSubmit: "UserPromptSubmit",
	PreToolUse:   "PreToolUse",
	PostToolUse:  "PostToolUse",
	Stop:         "Stop",
	PreCompact:   "PreCompact",
}

func tsPluginHarness(name, binary, installPkg, hookDir string) Harness {
	return Harness{
		Name: name, Binary: binary,
		InstallCmd: []string{"npm", "install", "-g", installPkg},
		APIFormat: Responses,
		EnvVars: map[string]string{
			"OPENAI_BASE_URL": "{{.BaseURL}}/v1",
		},
		APIKeyEnvVar:    "OPENAI_API_KEY",
		DefaultModel:    "openai/gpt-4o-mini",
		ToolCallName:    "read",
		ToolCallArgs:    `{"filePath":"README.md"}`,
		HookFormat:      TSPlugin,
		HookConfigDir:   hookDir,
		Events: Events{
			PromptSubmit: "experimental.chat.system.transform",
			PreToolUse:   "tool.execute.before",
			PostToolUse:  "tool.execute.after",
			Stop:         "session.idle",
		},
		PreserveHome:            true,
		NeedsGitRepo:            true,
		HeadlessCmd:             []string{binary, "run"},
		HeadlessModelArgs:       []string{"-m", "{{.Model}}", "--auto"},
		HooksInHeadless:         true,
		InteractiveCmd:          []string{binary, "--mini"},
		InteractiveArgs:         []string{"--auto", "-m", "{{.Model}}", "--prompt", "What is the project codename? Reply ONLY the codename."},
		InteractivePromptInArgs: true,
		HooksInInteractive:      true,
	}
}

var codexProviderArgs = []string{
	"-c", `model="{{.Model}}"`,
	"-c", `model_provider="mock"`,
	"-c", `model_providers.mock.name="Mock"`,
	"-c", `model_providers.mock.base_url="{{.BaseURL}}"`,
	"-c", `model_providers.mock.env_key="OPENAI_API_KEY"`,
	"-c", `model_providers.mock.wire_api="responses"`,
	"-c", `model_context_window=2048`,
}

var All = map[string]Harness{
	"claude": {
		Name: "claude", Binary: "claude",
		InstallCmd: []string{"npm", "install", "-g", "@anthropic-ai/claude-code"},
		APIFormat: Anthropic,
		EnvVars: map[string]string{
			"ANTHROPIC_BASE_URL":       "{{.BaseURL}}",
			"ANTHROPIC_AUTH_TOKEN":     "mock-auth-token",
			"CLAUDE_CODE_OAUTH_TOKEN":  "mock-oauth-token",
		},
		APIKeyEnvVar: "ANTHROPIC_API_KEY",
		DefaultModel: "claude-haiku-4-5-20251001",
		HookFormat:    JSONNested,
		HookConfigDir: ".claude",
		HookFileName:  "settings.json",
		HookWrapper:   `{"permissions":{"allow":["Bash(*)","Read(*)","Write(*)"]},"skipDangerousModePermissionPrompt":true,"hooks":%s}`,
		ConfigFiles: []ConfigFile{
			{Path: ".claude.json", Content: `{"projects":{"{{.RepoDir}}":{"hasTrustDialogAccepted":true}}}`},
			{Path: ".claude/settings.local.json", Content: `{"theme":"dark","hasCompletedOnboarding":true}`},
		},
		Events:             standardEvents,
		SkillsDir:          ".claude/skills",
		HeadlessCmd:         []string{"claude", "-p"},
		HeadlessModelArgs:   []string{"--model", "{{.Model}}", "--dangerously-skip-permissions", "--max-turns", "2"},
		PostHeadlessCmd: [][]string{{"-p", "--continue", "--dangerously-skip-permissions", "/compact"}},
		NeedsGitRepo:        true,
		HooksInHeadless:     true,
		InteractiveCmd:          []string{"claude"},
		InteractiveArgs:         []string{"--model", "{{.Model}}", "What is the project codename? Reply ONLY the codename."},
		InteractivePromptInArgs: true,
		ExitCommand:             "/exit",
		CompactCommand:          "/compact",
		HooksInInteractive:      true,
		OnboardingDismiss: []DismissAction{
			{Pattern: "theme"}, {Pattern: "Theme"}, {Pattern: "style"},
			{Pattern: "trust"}, {Pattern: "Trust"}, {Pattern: "onboarding"},
		},
	},
	"codex": {
		Name: "codex", Binary: "codex",
		InstallCmd: []string{"npm", "install", "-g", "@openai/codex"},
		PostInstall: [][]string{
			{"sh", "-c", "codex plugin marketplace add https://github.com/belt-sh/skills.git 2>/dev/null || true"},
			{"sh", "-c", "codex plugin add belt@belt-sh-skills 2>/dev/null || true"},
		},
		APIFormat: Responses,
		EnvVars: map[string]string{},
		APIKeyEnvVar:    "OPENAI_API_KEY",
		DefaultModel:    "gpt-4o-mini",
		ToolCallName:    "exec_command",
		ToolCallArgs:    `{"cmd":"cat README.md"}`,
		HookToolMatcher: "Bash",
		HookFormat:      JSONNested,
		HookConfigDir:   ".codex",
		HookFileName:    "hooks.json",
		Events:             standardEvents,
		SkillsDir:          ".agents/skills",
		HeadlessCmd:         []string{"codex", "exec"},
		HeadlessModelArgs: append([]string{
			"--dangerously-bypass-hook-trust",
			"--approve-for-me",
		}, codexProviderArgs...),
		PromptViaStdin:      true,
		NeedsGitRepo:        true,
		HooksInHeadless:     true,
		InteractiveCmd: []string{"codex"},
		InteractiveArgs: append(append([]string{
			"--dangerously-bypass-hook-trust",
			"-a", "never",
		}, codexProviderArgs...), "What is the project codename? Reply ONLY the codename."),
		InteractivePromptInArgs: true,
		ExitCommand:             "/exit",
		CompactCommand:          "/compact",
		HooksInInteractive:      true,
	},
	"copilot": {
		Name: "copilot", Binary: "copilot",
		InstallCmd: []string{"npm", "install", "-g", "@github/copilot"},
		APIFormat: OpenAI,
		EnvVars: map[string]string{
			"COPILOT_PROVIDER_BASE_URL": "{{.BaseURL}}",
			"COPILOT_MODEL":            "{{.Model}}",
		},
		APIKeyEnvVar: "COPILOT_PROVIDER_API_KEY",
		DefaultModel: "gpt-4o-mini",
		HookFormat:    JSONCopilot,
		HookConfigDir: ".copilot/hooks",
		Events: Events{
			PromptSubmit: "userPromptSubmitted",
			Stop:         "sessionEnd",
		},
		SkillsDir:          ".copilot/skills",
		HeadlessCmd:         []string{"copilot", "--prompt"},
		HooksInHeadless:     true,
		InteractiveCmd:          []string{"copilot", "-i", "What is the project codename? Reply ONLY the codename."},
		InteractivePromptInArgs: true,
		ExitCommand:             "/exit",
		HooksInInteractive:  true,
	},
	"grok": {
		Name: "grok", Binary: "grok",
		InstallCmd:     []string{"sh", "-c", "curl -fsSL https://x.ai/cli/install.sh | bash"},
		InstallBinDirs: []string{".grok/bin"},
		APIFormat:    Responses,
		ToolCallName: "read_file",
		ToolCallArgs: `{"target_file":"README.md"}`,
		ToolCallPath: "chat/completions",
		EnvVars: map[string]string{
			"GROK_CLI_CHAT_PROXY_BASE_URL": "{{.BaseURL}}",
			"GROK_XAI_API_BASE_URL":        "{{.BaseURL}}",
			"GROK_MODELS_BASE_URL":          "{{.BaseURL}}",
			"GROK_FEEDBACK_BASE_URL":        "{{.BaseURL}}",
			"GROK_TRACE_UPLOAD_URL":         "{{.BaseURL}}",
			"GROK_MANAGED_CONFIG_URL":       "{{.BaseURL}}",
			"GROK_CONVERSATIONS_BASE_URL":   "{{.BaseURL}}",
		},
		APIKeyEnvVar: "XAI_API_KEY",
		DefaultModel: "grok-3-mini",
		HookFormat:    JSONNested,
		HookConfigDir: ".grok/hooks",
		Events: Events{
			PromptSubmit: "UserPromptSubmit",
			PreToolUse:   "PreToolUse",
			PostToolUse:  "PostToolUse",
			Stop:         "Stop",
		},
		ConfigFiles: []ConfigFile{
			{Path: ".grok/auth.json", Content: `{"https://auth.x.ai::b1a00492-073a-47ea-816f-4c329264a828":{"key":"mock-test-token","auth_mode":"oidc","create_time":"2026-01-01T00:00:00Z","user_id":"mock-user","email":"mock@test.invalid","expires_at":"2030-01-01T00:00:00Z","refresh_token":"mock-refresh-token","oidc_issuer":"https://auth.x.ai","oidc_client_id":"b1a00492-073a-47ea-816f-4c329264a828","coding_data_retention_opt_out":false}}`},
			{Path: ".grok/trusted_folders.toml", Content: "[folders.\"{{.RepoDir}}\"]\ntrusted = true\ndecided_at = 1786997000\n"},
		},
		SkillsDir:           ".grok/skills",
		HeadlessCmd:          []string{"grok", "-p"},
		HooksInHeadless:      true,
		InteractiveCmd:       []string{"grok", "--always-approve"},
		ExitCommand:          "/exit",
		CompactCommand:       "/compact",
		HooksInInteractive:   true,
	},
	"pi": {
		Name: "pi", Binary: "pi",
		InstallCmd: []string{"npm", "install", "-g", "--ignore-scripts", "@earendil-works/pi-coding-agent"},
		APIFormat:    OpenAI,
		EnvVars: map[string]string{},
		APIKeyEnvVar: "OPENROUTER_API_KEY",
		DefaultModel: "openai/gpt-4o-mini",
		HookFormat:    TSExtension,
		HookConfigDir: ".pi/agent/extensions",
		Events:        Events{PromptSubmit: "before_agent_start", Stop: "agent_end"},
		HeadlessCmd:         []string{"pi", "-p"},
		HeadlessModelArgs:   []string{"--provider", "openrouter", "--model", "{{.Model}}", "--no-session"},
		HooksInHeadless:     true,
		InteractiveCmd:      []string{"pi"},
		ExitCommand:         "/exit",
		HooksInInteractive:  true,
	},
	"hermes": {
		Name: "hermes", Binary: "hermes",
		InstallCmd:     []string{"pip", "install", "--break-system-packages", "hermes-agent"},
		InstallBinDirs: []string{".local/bin"},
		APIFormat:    OpenAI,
		ToolCallName: "read_file",
		ToolCallArgs: `{"path":"README.md"}`,
		EnvVars: map[string]string{
			"OPENROUTER_BASE_URL": "{{.BaseURL}}/v1",
		},
		APIKeyEnvVar: "OPENROUTER_API_KEY",
		DefaultModel: "openai/gpt-4o-mini",
		HookFormat:    YAML,
		HookConfigDir: ".hermes",
		Events: Events{
			PromptSubmit: "pre_llm_call",
			PreToolUse:   "pre_tool_call",
			PostToolUse:  "post_tool_call",
			Stop:         "on_session_end",
		},
		ConfigFiles: []ConfigFile{
			{Path: ".hermes/.env", Content: "OPENROUTER_API_KEY={{.APIKey}}\n"},
			{Path: ".hermes/config.yaml", Content: "model:\n  default: {{.Model}}\n  provider: openrouter\n  base_url: {{.BaseURL}}/v1\nhooks: {}\nhooks_auto_accept: true\n"},
		},
		HeadlessCmd:          []string{"hermes", "chat", "-q"},
		HeadlessModelArgs:    []string{"-m", "{{.Model}}", "--provider", "openrouter", "--accept-hooks"},
		HooksInHeadless:      true,
		InteractiveCmd:       []string{"hermes"},
		ExitCommand:          "/exit",
		HooksInInteractive:   true,
	},
	"kilo": tsPluginHarness("kilo", "kilo", "@kilocode/cli", ".kilo/plugins"),
	"kimi": {
		Name: "kimi", Binary: "kimi",
		InstallCmd: []string{"npm", "install", "-g", "@moonshot-ai/kimi-code"},
		APIFormat: OpenAI,
		EnvVars: map[string]string{
			"KIMI_CODE_BASE_URL": "{{.BaseURL}}/coding/v1",
		},
		APIKeyEnvVar:    "",
		DefaultModel:    "gpt-4o-mini",
		ToolCallName:    "Read",
		ToolCallArgs:    `{"path":"{{.RepoDir}}/README.md"}`,
		HookFormat:      TOML,
		HookConfigDir:   ".kimi-code",
		Events: Events{
			SessionStart: "SessionStart",
			PromptSubmit: "UserPromptSubmit",
			PreToolUse:   "PreToolUse",
			PostToolUse:  "PostToolUse",
			Stop:         "Stop",
		},
		ConfigFiles: []ConfigFile{
			{Path: ".kimi-code/config.toml", Content: "default_model = \"mock\"\n\n[providers.mock]\ntype = \"openai\"\napi_key = \"mock-key\"\nbase_url = \"{{.BaseURL}}/v1\"\n\n[models.mock]\nprovider = \"mock\"\nmodel = \"{{.Model}}\"\nmax_context_size = 128000\nmax_output_size = 4096\n"},
			{Path: ".kimi-code/credentials/kimi-code-env-{{.TokenHash16}}.json", Content: `{"access_token":"mock-kimi-token","refresh_token":"mock-refresh","expires_at":99999999999,"scope":"","token_type":"Bearer","expires_in":86400}`},
		},
		NeedsGitRepo:        true,
		HeadlessCmd:         []string{"kimi", "-p"},
		HooksInHeadless:     true,
		InteractiveCmd:      []string{"kimi"},
		ExitCommand:         "/exit",
		HooksInInteractive:  true,
		OnboardingDismiss: []DismissAction{{Pattern: "Don't trust", SendUp: true}},
	},
	"goose": {
		Name: "goose", Binary: "goose",
		InstallCmd:     []string{"sh", "-c", "mkdir -p $HOME/.local/bin && curl -fsSL https://github.com/aaif-goose/goose/releases/download/stable/goose-x86_64-unknown-linux-gnu.tar.bz2 | tar -xj --strip-components=0 -C $HOME/.local/bin"},
		InstallBinDirs: []string{".local/bin"},
		APIFormat:    OpenAI,
		EnvVars: map[string]string{
			"GOOSE_PROVIDER":               "mock",
			"GOOSE_MODEL":                  "{{.Model}}",
			"GOOSE_MODE":                   "auto",
			"GOOSE_DISABLE_SESSION_NAMING": "true",
			"OPENAI_API_KEY":               "mock-key",
		},
		APIKeyEnvVar: "",
		DefaultModel: "gpt-4o-mini",
		ToolCallName: "shell",
		ToolCallArgs: `{"command":"cat README.md"}`,
		ConfigFiles: []ConfigFile{
			{Path: ".config/goose/custom_providers/mock.json", Content: `{"name":"mock","engine":"openai","display_name":"Mock","api_key_env":"OPENAI_API_KEY","base_url":"{{.BaseURL}}/v1/chat/completions","models":[{"name":"gpt-4o-mini","context_limit":128000}],"supports_streaming":true,"requires_auth":true}`},
			{Path: ".agents/plugins/belt-test/plugin.json", Content: `{"name":"belt-test","version":"1.0.0","description":"Belt harness test hooks"}`},
		},
		HookFormat:    JSONNested,
		HookConfigDir: ".agents/plugins/belt-test/hooks",
		HookFileName:  "hooks.json",
		Events: Events{
			SessionStart: "SessionStart",
			PromptSubmit: "UserPromptSubmit",
			PreToolUse:   "PreToolUse",
			PostToolUse:  "PostToolUse",
			Stop:         "Stop",
		},
		NeedsGitRepo:    true,
		HeadlessCmd:     []string{"goose", "run", "-t"},
		HooksInHeadless:  true,
		InteractiveCmd:   []string{"goose"},
		ExitCommand:      "/exit",
		HooksInInteractive: true,
	},
	"gemini": {
		Name: "gemini", Binary: "gemini",
		InstallCmd: []string{"npm", "install", "-g", "@google/gemini-cli"},
		APIFormat: Gemini,
		EnvVars: map[string]string{
			"GOOGLE_GEMINI_BASE_URL":     "{{.BaseURL}}",
			"GEMINI_CLI_TRUST_WORKSPACE": "true",
			"GEMINI_API_KEY":             "mock-key",
		},
		APIKeyEnvVar: "",
		DefaultModel: "gemini-2.5-flash",
		ToolCallName: "write_file",
		ToolCallArgs: `{"file_path":"{{.RepoDir}}/test-output.txt","content":"test"}`,
		HookFormat:    JSONNested,
		HookConfigDir: ".gemini",
		HookFileName:  "settings.json",
		HookWrapper:   `{"baseUrl":"{{.BaseURL}}","security":{"auth":{"selectedType":"gemini-api-key","useExternal":true}},"hooks":%s}`,
		Events: Events{
			SessionStart: "SessionStart",
			PromptSubmit: "BeforeAgent",
			PreToolUse:   "BeforeTool",
			PostToolUse:  "AfterTool",
			Stop:         "SessionEnd",
			PreCompact:   "PreCompress",
		},
		NeedsGitRepo:        true,
		HeadlessCmd:         []string{"gemini", "-p"},
		HeadlessModelArgs:   []string{"--yolo"},
		HooksInHeadless:     true,
		InteractiveCmd:          []string{"gemini", "--yolo", "-m", "gemini-2.5-flash", "-i", "What is the project codename? Reply ONLY the codename."},
		InteractivePromptInArgs: true,
		SlowInput:               true,
		ExitCommand:             "/exit",
		CompactCommand:          "/compress",
		HooksInInteractive:      true,
	},
	"qwen": {
		Name: "qwen", Binary: "qwen",
		InstallCmd: []string{"npm", "install", "-g", "@qwen-code/qwen-code"},
		APIFormat: OpenAI,
		EnvVars: map[string]string{
			"OPENAI_BASE_URL": "{{.BaseURL}}/v1",
		},
		APIKeyEnvVar:    "OPENAI_API_KEY",
		DefaultModel:    "gpt-4o-mini",
		ToolCallName:    "read_file",
		ToolCallArgs:    `{"file_path":"{{.RepoDir}}/README.md"}`,
		HookFormat:    JSONNested,
		HookConfigDir: ".qwen",
		HookFileName:  "settings.json",
		HookWrapper:   `{"permissions":{"allow":["Bash(*)","Read(*)","Write(*)"]},"hooks":%s}`,
		Events:              standardEvents,
		NeedsGitRepo:        true,
		HeadlessCmd:         []string{"qwen", "-p"},
		HeadlessModelArgs:   []string{"--model", "{{.Model}}", "--yolo", "--auth-type", "openai"},
		HooksInHeadless:     true,
		InteractiveCmd:          []string{"qwen"},
		InteractiveArgs:         []string{"--model", "{{.Model}}", "--yolo", "--auth-type", "openai"},
		SlowInput:               true,
		PostHeadlessCmd:         [][]string{{"-p", "--continue", "--yolo", "--auth-type", "openai", "/compress"}},
		ExitCommand:             "/exit",
		CompactCommand:          "/compress",
		HooksInInteractive:      true,
	},
	"opencode": tsPluginHarness("opencode", "opencode", "opencode-ai", ".opencode/plugins"),
	"droid": {
		Name: "droid", Binary: "droid",
		InstallCmd: []string{"npm", "install", "-g", "droid"},
		APIFormat: OpenAI,
		EnvVars: map[string]string{
			"FACTORY_API_BASE_URL": "{{.BaseURL}}",
			"FACTORY_API_KEY":      "fk-mock-key-0123456789abcdef0123",
			"FACTORY_DISABLE_KEYRING": "1",
		},
		APIKeyEnvVar:    "",
		DefaultModel:    "mock-model",
		ToolCallName:    "Read",
		ToolCallArgs:    `{"file_path":"{{.RepoDir}}/README.md"}`,
		HookFormat:      JSONNested,
		HookConfigDir:   ".factory",
		HookFileName:    "hooks.json",
		HookWrapper:     "%s",
		Events:          standardEvents,
		ConfigFiles: []ConfigFile{
			{Path: ".factory/settings.json", Content: `{"customModels":[{"model":"mock-model","displayName":"Mock","baseUrl":"{{.BaseURL}}/v1","apiKey":"mock-key","provider":"openai","maxOutputTokens":4096}]}`},
		},
		SkillsDir:       ".factory-plugin/skills",
		NeedsGitRepo:    true,
		HeadlessCmd:     []string{"droid", "exec"},
		HeadlessModelArgs: []string{"--auto", "high", "-m", "{{.Model}}", "-o", "stream-jsonrpc"},
		PostHeadlessCmd:   [][]string{{"exec", "--session-id", "{{.SessionID}}", "--auto", "high", "-m", "{{.Model}}", "/compact"}},
		HooksInHeadless:  true,
		InteractiveCmd:   []string{"droid"},
		InteractiveArgs:  []string{"--auto", "high", "-m", "{{.Model}}"},
		ExitCommand:      "/exit",
		CompactCommand:   "/compact",
		HooksInInteractive:    true,
	},
}

// Investigated but not added:
//
// Amp (@ampcode/cli) — Uses Rivet WebSocket protocol (/actors endpoint).
//   The CLI is a thin client; the agent loop runs server-side on ampcode.com.
//   Cannot mock with an HTTP LLM endpoint. Would need a full Rivet actor server.
//   Has TypeScript plugins via amp.on() with 5 events (tool.call, tool.result,
//   agent.start, agent.end, session.start). Headless: amp -x "prompt".
//
// Kiro (kirodotdev/Kiro) — Amazon-backed, 10 hook events (UserPromptSubmit,
//   PreToolUse, PostToolUse, SessionStart, Stop, PreTaskExec, PostTaskExec,
//   PostFileSave, PostFileCreate, PostFileDelete). JSON v1 format in
//   .kiro/hooks/<id>.json. No BYOK — uses Bedrock internally, no custom
//   endpoint support. Open feature request: github.com/kirodotdev/Kiro/issues/695.
//   Headless: kiro-cli chat --no-interactive "prompt". Auth: KIRO_API_KEY.
//
// Aider (aider-chat on PyPI) — No hook/plugin system. Has --lint-cmd and
//   --test-cmd post-edit hooks only. BYOK via OPENAI_API_BASE + OPENAI_API_KEY.
//   Headless: aider --message "prompt" --yes-always.
//
// Cursor / Windsurf / Cline / Roo — IDE extensions only, no standalone CLI.
//   Cursor has JSONFlat hook format but runs inside VS Code.
//
// Remaining skip investigations (5 skips across 3 harnesses, 245/5):
//
// Codex PreCompact (2 skips, H+I): Headless: /compact is TUI-only slash
//   command (slash_dispatch.rs). codex exec resume --last exists but treats
//   /compact as user message, not slash command. Interactive: TUI /compact
//   calls run_pre_compact_hooks() but doesn't fire hooks even with multiple
//   warmup turns (tested 3 warmups, 7 server requests). capture_step_context()
//   appears to require conditions the mock provider can't satisfy.
//   Session files: ~/.codex/sessions/YYYY/MM/DD/rollout-<ts>-<uuid>.jsonl.
//
// Droid PreCompact (2 skips, H+I): Headless: exec --session-id treats
//   /compact as user message, not slash command. Interactive: /compact sent
//   via SendLine but doesn't trigger PreCompact hooks. Context may be too
//   short for compaction threshold.
//
// Copilot interactive prompt (1 skip, I only): Not Ink — fully custom
//   React terminal renderer (S6 class). PTY SendLine: ICRNL race
//   converts \r→\n before raw mode set (submit only triggers on \r).
//   Paste coalescing (pSt function) + React batching means submit fires
//   with empty/stale buffer. -i flag: fire-and-forget async useEffect,
//   hook dispatch runs but bash command hasn't completed when process
//   exits. sessionEnd fires (later in lifecycle, after hooks loaded).
//
// Fixed (session 1):
//   Codex PreToolUse + PostToolUse (3 skips → 0): Hook matcher mismatch.
//     exec_command exposes as "Bash" via HookToolName::bash() in
//     codex-rs/core/src/tools/hook_names.rs. Added HookToolMatcher
//     field to decouple API tool name from hook matcher.
//   Gemini interactive (4 skips → 0): Two fixes. (1) selectedType:
//     "gemini-api-key" (was "gateway") with GEMINI_API_KEY env var.
//     GATEWAY routed through flash-lite planning model (no tools).
//     gemini-api-key gives full agent model access. (2) -m gemini-2.5-flash
//     skips flash-lite routing, uses streamGenerateContent with full
//     functionDeclarations (56KB+ tool payload).
//   Qwen interactive PreCompact (1 skip → 0): Positional prompt arg made
//     qwen run in headless mode. Removed prompt from InteractiveArgs,
//     now uses SendLine + SlowInput in actual TUI mode.
//
// Fixed (session 2):
//   Droid interactive PreToolUse + PostToolUse (2 skips → 0): TUI routes
//     LLM requests through Factory API proxy at /api/llm/a/v1/messages
//     (Anthropic Messages format), not /v1/chat/completions. Mock server
//     had no handler for this path — requests hit catch-all, got empty
//     200 responses. Added Factory proxy path handlers. Droid TUI sends
//     claude-haiku (tools=0, planning) then claude-opus (tools=18, agent).
//     hasTools check correctly skips planning request, fires tool call on
//     agent request.
//
// TUI technology map:
//   Ink (React): claude, grok, droid, kimi, pi, qwen, gemini, opencode, kilo
//   Custom React renderer: copilot (S6 class, not Ink)
//   crossterm (Rust): codex (Bun-bundled codex-rs, startup input quarantine)
//   Ratatui (Rust): goose (crossterm works, no quarantine)
//   prompt_toolkit (Python): hermes
//
// PTY input compatibility:
//   SendLine works: claude, grok, goose, droid, kimi, pi, hermes, qwen
//   SendLine + SlowInput: gemini (anti-paste 30ms), qwen (/compress)
//   Prompt-in-args only: codex (startup quarantine), copilot (ICRNL race)
//   --prompt flag: opencode, kilo (TSPlugin via --prompt flag)
