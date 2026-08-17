package harness

// HookFormat describes how a harness expects hook configuration.
type HookFormat int

const (
	JSONNested   HookFormat = iota // Claude, Codex, Grok, Droid, Qoder
	JSONFlat                       // Cursor, Windsurf
	JSONCopilot                    // Copilot v1 format (version field, bash field)
	TOML                           // Kimi
	YAML                           // Hermes
	TSExtension                    // Pi
	TSPlugin                       // OpenCode, Kilo
)

// APIFormat describes what LLM API protocol the harness speaks.
type APIFormat int

const (
	OpenAI    APIFormat = iota // /v1/chat/completions
	Responses                  // /v1/responses (Codex)
	Anthropic                  // /v1/messages
	Google                     // /v1/models/*/generateContent
)

// EventCase describes the naming convention for hook events.
type EventCase int

const (
	PascalCase EventCase = iota // SessionStart
	CamelCase                   // sessionStart
	SnakeCase                   // session_start
)

// Harness describes a coding agent CLI and how belt integrates with it.
type Harness struct {
	Name    string
	Binary  string   // CLI binary name
	Install []string // install commands to try in order

	// API
	APIFormat       APIFormat
	EndpointEnvVar  string // env var for custom base URL (empty = not supported)
	APIKeyEnvVar    string // env var for API key
	ModelFlag       string // CLI flag or config for model selection
	DefaultModel    string // model ID to use in tests

	// Hooks
	HookFormat    HookFormat
	EventCase     EventCase
	HookConfigDir string // where hook config goes (relative to $HOME)
	Events        Events // event name mapping

	// Skills
	SkillsDir string // where SKILL.md files go (relative to $HOME)

	// Headless (-p) mode
	HeadlessCmd     []string // command to run a single prompt headlessly
	HeadlessExtraFlags []string // extra flags for headless mode
	NeedsGitRepo    bool     // must be run inside a git repo
	NeedsNonRoot    bool     // refuses to run as root
	HooksInHeadless bool     // whether hooks fire in headless mode

	// Interactive (PTY/TUI) mode
	InteractiveCmd  []string // command to start interactive session
	InteractiveExtraFlags []string
	ExitCommand     string   // slash command or keystroke to exit (e.g. "/exit", "Ctrl+C")
	HooksInInteractive bool // whether hooks fire in interactive mode (usually true)
	NeedsTrustSetup bool    // needs pre-trust config for hooks to fire

	// Context injection
	InjectionMethod string // how hook output reaches the agent context
	CanInject       bool   // whether context injection is possible at all
}

// Events maps belt behaviors to harness-specific event names.
type Events struct {
	SessionStart string
	PromptSubmit string // the pre-prompt event for suggest injection
	PreToolUse   string
	PostToolUse  string
	Stop         string
	PreCompact   string
	SessionEnd   string
}
