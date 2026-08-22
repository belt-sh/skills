package harness

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// DetectResult describes a detected harness installation.
type DetectResult struct {
	Name       string
	Binary     string // resolved binary path (empty if not in PATH)
	Version    string // version string from --version (empty if unavailable)
	ConfigDir  string // detected config directory (empty if not found)
	HooksPath  string // where hooks should be written (user-level)
	Installed  bool   // binary found via any strategy
	Configured bool   // config dir exists (harness has been used)
}

// configDirs maps harness names to their config directory relative to $HOME.
var configDirs = map[string][]string{
	"claude":   {".claude"},
	"codex":    {".codex"},
	"copilot":  {".copilot"},
	"droid":    {".factory"},
	"gemini":   {".gemini"},
	"goose":    {".config/goose"},
	"grok":     {".grok"},
	"hermes":   {".hermes"},
	"kilo":     {".config/kilo"},
	"kimi":     {".kimi-code"},
	"opencode": {".config/opencode"},
	"pi":       {".pi"},
	"qwen":     {".qwen"},
}

// hooksTargets maps harness names to the user-level hook file path relative to $HOME.
var hooksTargets = map[string]string{
	"claude":   ".claude/settings.json",
	"codex":    ".codex/hooks.json",
	"copilot":  ".copilot/hooks/belt.json",
	"droid":    ".factory/hooks.json",
	"gemini":   ".gemini/settings.json",
	"goose":    ".agents/plugins/belt/hooks/hooks.json",
	"grok":     ".grok/hooks/belt.json",
	"hermes":   ".hermes/config.yaml",
	"kilo":     ".config/kilo/plugin/belt.ts",
	"kimi":     ".kimi-code/config.toml",
	"opencode": ".config/opencode/plugins/belt.ts",
	"pi":       ".pi/agent/extensions/belt.ts",
	"qwen":     ".qwen/settings.json",
}

// knownBinPaths are extra directories to search beyond PATH.
var knownBinPaths = []string{
	".local/bin",
	".npm-global/bin",
	".grok/bin",
	".cargo/bin",
	".nvm/versions/node", // contains nested bin/ dirs
}

// npmPackages maps harness names to their npm package names.
var npmPackages = map[string]string{
	"claude":   "@anthropic-ai/claude-code",
	"codex":    "@openai/codex",
	"copilot":  "@github/copilot",
	"droid":    "droid",
	"gemini":   "@google/gemini-cli",
	"kilo":     "@kilocode/cli",
	"kimi":     "@moonshot-ai/kimi-code",
	"opencode": "opencode-ai",
	"pi":       "@earendil-works/pi-coding-agent",
	"qwen":     "@qwen-code/qwen-code",
}

// DetectAll runs layered detection for every known harness.
func DetectAll() []DetectResult {
	home, _ := os.UserHomeDir()
	var results []DetectResult

	for name, h := range All {
		r := detect(name, h.Binary, home)
		results = append(results, r)
	}
	return results
}

// DetectOne runs layered detection for a single harness.
func DetectOne(name string) DetectResult {
	h, ok := All[name]
	if !ok {
		return DetectResult{Name: name}
	}
	home, _ := os.UserHomeDir()
	return detect(name, h.Binary, home)
}

func detect(name, binary, home string) DetectResult {
	r := DetectResult{Name: name}

	// Layer 1: config dir exists (most reliable — means harness was used here)
	if dirs, ok := configDirs[name]; ok {
		for _, d := range dirs {
			full := filepath.Join(home, d)
			if info, err := os.Stat(full); err == nil && info.IsDir() {
				r.ConfigDir = full
				r.Configured = true
				break
			}
		}
	}

	// Layer 2: exec.LookPath (cross-platform binary detection)
	if path, err := exec.LookPath(binary); err == nil {
		r.Binary = path
		r.Installed = true
	}

	// Layer 3: known install paths (catches binaries not in PATH)
	if !r.Installed && home != "" {
		for _, binDir := range knownBinPaths {
			candidate := filepath.Join(home, binDir, binary)
			if _, err := os.Stat(candidate); err == nil {
				r.Binary = candidate
				r.Installed = true
				break
			}
		}
	}

	// Layer 4: npm global list (catches npm-installed packages regardless of PATH)
	if !r.Installed {
		if pkg, ok := npmPackages[name]; ok {
			if npmHasGlobal(pkg) {
				r.Installed = true
			}
		}
	}

	// Version detection (only if we found the binary)
	if r.Binary != "" {
		r.Version = getVersion(r.Binary)
	}

	// Hook target path
	if target, ok := hooksTargets[name]; ok {
		r.HooksPath = filepath.Join(home, target)
	}

	return r
}

func getVersion(binary string) string {
	out, err := exec.Command(binary, "--version").CombinedOutput()
	if err != nil {
		return ""
	}
	v := strings.TrimSpace(string(out))
	if i := strings.IndexByte(v, '\n'); i > 0 {
		v = v[:i]
	}
	return v
}

var npmGlobalCache map[string]bool

func npmHasGlobal(pkg string) bool {
	if npmGlobalCache == nil {
		npmGlobalCache = make(map[string]bool)
		out, err := exec.Command("npm", "list", "-g", "--depth=0", "--json").Output()
		if err != nil {
			return false
		}
		var result struct {
			Dependencies map[string]any `json:"dependencies"`
		}
		if json.Unmarshal(out, &result) == nil {
			for k := range result.Dependencies {
				npmGlobalCache[k] = true
			}
		}
	}
	return npmGlobalCache[pkg]
}
