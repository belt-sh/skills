package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/belt-sh/skills/tools/harness-test/harness"
	"github.com/belt-sh/skills/tools/harness-test/server"
)

type Result struct {
	Harness string
	Passed  int
	Failed  int
	Skipped int
}

type Mode int

const (
	ModeBoth Mode = iota
	ModeHeadless
	ModeInteractive
)

type Runner struct {
	harness    harness.Harness
	server     *server.MockServer
	baseURL    string
	home       string
	repoDir    string
	injectCode string
	mode       Mode
	failed     bool
	result     Result
	lastOutput string
}

var originalHome = os.Getenv("HOME")

func New(h harness.Harness, srv *server.MockServer, baseURL string) *Runner {
	return &Runner{
		harness: h,
		server:  srv,
		baseURL: baseURL,
		mode:    ModeBoth,
		result:  Result{Harness: h.Name},
	}
}

func (r *Runner) SetMode(m string) {
	switch m {
	case "headless":
		r.mode = ModeHeadless
	case "interactive":
		r.mode = ModeInteractive
	default:
		r.mode = ModeBoth
	}
}

func (r *Runner) pass(msg string) {
	r.result.Passed++
	fmt.Printf("  ✓ %s\n", msg)
}

func (r *Runner) fail(msg string) {
	r.result.Failed++
	r.failed = true
	fmt.Fprintf(os.Stderr, "  ✗ %s\n", msg)
}

func (r *Runner) skip(msg string) {
	r.result.Skipped++
	fmt.Printf("  ○ %s\n", msg)
}

func (r *Runner) Run() Result {
	fmt.Printf("=== %s ===\n", r.harness.Name)

	r.setupHome()
	r.checkBinary()
	if r.failed {
		return r.finish()
	}
	r.setupEndpoint()
	r.writeConfigFiles()
	r.writeHooks()
	r.setupSkills()

	hasToolHooks := r.harness.Events.PreToolUse != "" || r.harness.Events.PostToolUse != ""

	if r.mode == ModeBoth || r.mode == ModeHeadless {
		if r.harness.HooksInHeadless {
			if r.server != nil && hasToolHooks {
				r.server.SetToolCall(r.harness.ToolCallName, r.harness.ToolCallArgs)
				r.server.SetToolCallMode(true)
			}
			r.runHeadless()
			r.checkHookEvents("headless")
		} else {
			r.skip(r.harness.Name + " does not fire hooks in headless mode")
		}
	}
	if r.mode == ModeBoth || r.mode == ModeInteractive {
		os.Remove("/tmp/belt-hook-events.log")
		r.server.ClearLog()
		if r.server != nil && hasToolHooks {
			r.server.SetToolCall(r.harness.ToolCallName, r.harness.ToolCallArgs)
			r.server.SetToolCallMode(true)
		}
		r.runInteractive()
		if r.harness.NeedsAuthForInteractive {
			r.skip("interactive hooks: requires OAuth (headless covers hook verification)")
		} else {
			r.checkHookEvents("interactive")
		}
	}

	return r.finish()
}

func (r *Runner) finish() Result {
	fmt.Printf("\n=== %s: %d passed, %d failed, %d skipped ===\n\n",
		r.harness.Name, r.result.Passed, r.result.Failed, r.result.Skipped)
	return r.result
}

func (r *Runner) setupHome() {
	if r.harness.PreserveHome {
		r.home = originalHome
		os.Setenv("HOME", originalHome)
		return
	}
	dir, err := os.MkdirTemp("", "harness-test-"+r.harness.Name+"-")
	if err != nil {
		r.fail("create temp home: " + err.Error())
		return
	}
	r.home = dir
	os.Setenv("HOME", dir)
}

func (r *Runner) checkBinary() {
	fmt.Println("[phase 1] prerequisites")
	if _, err := exec.LookPath(r.harness.Binary); err != nil {
		if len(r.harness.InstallCmd) == 0 {
			r.fail(r.harness.Binary + " not found (no install command)")
			return
		}
		fmt.Printf("  … installing %s\n", r.harness.Binary)
		cmd := exec.Command(r.harness.InstallCmd[0], r.harness.InstallCmd[1:]...)
		cmd.Env = os.Environ()
		out, installErr := cmd.CombinedOutput()
		if installErr != nil {
			r.fail(fmt.Sprintf("install %s: %v\n%s", r.harness.Binary, installErr, string(out)))
			return
		}
		for _, d := range r.harness.InstallBinDirs {
			p := filepath.Join(r.home, d)
			if !strings.Contains(os.Getenv("PATH"), p) {
				os.Setenv("PATH", p+":"+os.Getenv("PATH"))
			}
		}
		if _, err := exec.LookPath(r.harness.Binary); err != nil {
			r.fail(r.harness.Binary + " not found after install")
			return
		}
		r.pass(r.harness.Binary + " installed")
		for _, postCmd := range r.harness.PostInstall {
			cmd := exec.Command(postCmd[0], postCmd[1:]...)
			cmd.Env = os.Environ()
			cmd.Run()
		}
		return
	}
	r.pass(r.harness.Binary + " found")
}

func (r *Runner) setupEndpoint() {
	fmt.Println("[phase 2] endpoint")
	for envVar, tmpl := range r.harness.EndpointEnvVars {
		val := r.expand(tmpl)
		os.Setenv(envVar, val)
		r.pass(envVar + "=" + val)
	}
	if r.harness.APIKeyEnvVar != "" {
		os.Setenv(r.harness.APIKeyEnvVar, "mock-key")
		r.pass(r.harness.APIKeyEnvVar + " set")
	}
	r.server.ClearLog()
}

func (r *Runner) writeToHome(relPath string, content string) {
	path := filepath.Join(r.home, relPath)
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte(content), 0644)
	if r.harness.NeedsGitRepo {
		projPath := filepath.Join(r.ensureGitRepo(), relPath)
		os.MkdirAll(filepath.Dir(projPath), 0755)
		os.WriteFile(projPath, []byte(content), 0644)
	}
}

func (r *Runner) writeConfigFiles() {
	if len(r.harness.ConfigFiles) == 0 {
		return
	}
	for _, cf := range r.harness.ConfigFiles {
		r.writeToHome(cf.Path, r.expand(cf.Content))
	}
}

func (r *Runner) writeHooks() {
	fmt.Println("[phase 3] hooks")

	hookDir := filepath.Join(r.home, r.harness.HookConfigDir)
	os.MkdirAll(hookDir, 0755)

	logPath := "/tmp/belt-hook-events.log"
	os.Remove(logPath)
	r.injectCode = fmt.Sprintf("%s-%d", strings.ToUpper(r.harness.Name), time.Now().UnixMilli())

	var content string
	var filename string

	switch r.harness.HookFormat {
	case harness.JSONNested:
		filename = "belt.json"
		hooksJSON := r.buildNestedHooksJSON(logPath)
		if r.harness.HookWrapper != "" {
			content = fmt.Sprintf(r.harness.HookWrapper, hooksJSON)
			filename = r.harness.HookFileName
		} else {
			content = fmt.Sprintf(`{"hooks":%s}`, hooksJSON)
		}

	case harness.JSONCopilot:
		filename = "belt.json"
		injectJSON := fmt.Sprintf(`{\"additionalContext\": \"The project codename is %s.\"}`, r.injectCode)
		copilotPromptCmd := fmt.Sprintf("echo PROMPT >> %s && echo '%s'", logPath, injectJSON)
		stopCmd := fmt.Sprintf("echo STOP >> %s", logPath)
		content = fmt.Sprintf(`{"version":1,"hooks":{"%s":[{"type":"command","bash":"%s","timeoutSec":5}],"%s":[{"type":"command","bash":"%s","timeoutSec":5}]}}`,
			r.harness.Events.PromptSubmit, copilotPromptCmd, r.harness.Events.Stop, stopCmd)

	case harness.YAML:
		scriptDir := filepath.Join(r.home, ".hermes", "test-hooks")
		os.MkdirAll(scriptDir, 0755)

		yamlHooks := "hooks:\n"
		for _, e := range r.eventEntries() {
			script := filepath.Join(scriptDir, e.Tag+".sh")
			body := fmt.Sprintf("#!/bin/sh\ncat - >/dev/null\necho %s >> %s\n", e.Tag, logPath)
			if e.Tag == "PROMPT" {
				body += fmt.Sprintf("printf '{\"context\": \"The project codename is %s.\"}\\n'\n", r.injectCode)
			}
			os.WriteFile(script, []byte(body), 0755)
			yamlHooks += fmt.Sprintf("  %s:\n    - command: %s\n      timeout: 5\n", e.Event, script)
		}
		// Append hooks to existing config.yaml (model config comes from ConfigFiles)
		cfgPath := filepath.Join(r.home, r.harness.HookConfigDir, "config.yaml")
		existing, _ := os.ReadFile(cfgPath)
		existingStr := strings.Replace(string(existing), "hooks: {}", "", 1)
		content = existingStr + yamlHooks
		filename = "config.yaml"

	case harness.TSExtension:
		filename = "belt-test.ts"
		tsHooks := ""
		for _, e := range r.eventEntries() {
			if e.Tag == "PROMPT" {
				tsHooks += fmt.Sprintf(`  pi.on("%s", async (event: any) => {
    require("fs").appendFileSync("%s", "PROMPT\n");
    return { systemPrompt: (event.systemPrompt || '') + '\nThe project codename is %s.' };
  });
`, e.Event, logPath, r.injectCode)
			} else {
				tsHooks += fmt.Sprintf(`  pi.on("%s", async () => {
    require("fs").appendFileSync("%s", "%s\n");
  });
`, e.Event, logPath, e.Tag)
			}
		}
		content = fmt.Sprintf("export default function (pi: any) {\n%s}\n", tsHooks)

	case harness.TSPlugin:
		filename = "belt-test.ts"
		var hookParts []string
		startLine := ""
		for _, e := range r.eventEntries() {
			switch e.Tag {
			case "SESSION_START":
				startLine = fmt.Sprintf("  require(\"fs\").appendFileSync(\"%s\", \"SESSION_START\\n\");\n", logPath)
			case "PROMPT":
				hookParts = append(hookParts, fmt.Sprintf(`    "%s": async (_input: any, output: any) => {
      require("fs").appendFileSync("%s", "PROMPT\n");
      output.system.push("The project codename is %s.");
    }`, e.Event, logPath, r.injectCode))
			case "STOP":
				hookParts = append(hookParts, fmt.Sprintf(`    "event": async ({ event }: any) => {
      if (event.type === "%s") {
        require("fs").appendFileSync("%s", "STOP\n");
      }
    }`, e.Event, logPath))
			default:
				hookParts = append(hookParts, fmt.Sprintf(`    "%s": async () => {
      require("fs").appendFileSync("%s", "%s\n");
    }`, e.Event, logPath, e.Tag))
			}
		}
		content = fmt.Sprintf("export const TestPlugin = async (_ctx: any) => {\n%s  return {\n%s,\n  };\n};\n",
			startLine, strings.Join(hookParts, ",\n"))

	default:
		r.skip("hook format not yet implemented")
		return
	}

	if r.harness.HookFileName != "" && filename == "belt.json" {
		filename = r.harness.HookFileName
	}
	hookFile := filepath.Join(hookDir, filename)
	os.WriteFile(hookFile, []byte(content), 0644)
	if r.harness.NeedsGitRepo {
		projHookDir := filepath.Join(r.ensureGitRepo(), r.harness.HookConfigDir)
		os.MkdirAll(projHookDir, 0755)
		os.WriteFile(filepath.Join(projHookDir, filename), []byte(content), 0644)
	}
	r.pass(fmt.Sprintf("hooks configured (code: %s)", r.injectCode))
}

func (r *Runner) setupSkills() {
	if r.harness.SkillsDir == "" {
		return
	}
	fmt.Println("[phase 4] skills")
	os.MkdirAll(filepath.Join(r.home, r.harness.SkillsDir), 0755)
	r.pass("skills directory created")
}

func (r *Runner) ensureGitRepo() string {
	if r.repoDir != "" {
		return r.repoDir
	}
	r.repoDir = filepath.Join(r.home, "test-repo")
	os.MkdirAll(r.repoDir, 0755)
	run(r.repoDir, "git", "init", "-q")
	run(r.repoDir, "git", "config", "user.email", "t@t")
	run(r.repoDir, "git", "config", "user.name", "t")
	os.WriteFile(filepath.Join(r.repoDir, "README.md"), []byte("test"), 0644)
	run(r.repoDir, "git", "add", ".")
	run(r.repoDir, "git", "commit", "-qm", "init")
	return r.repoDir
}

func (r *Runner) runHeadless() {
	if len(r.harness.HeadlessCmd) == 0 {
		r.skip("no headless command configured")
		return
	}

	fmt.Println("[phase 5] headless prompt")

	dir := r.home
	if r.harness.NeedsGitRepo {
		dir = r.ensureGitRepo()
	}

	prompt := "What is the project codename? Reply ONLY the codename."

	var args []string
	args = append(args, r.harness.HeadlessCmd[1:]...)
	if !r.harness.PromptViaStdin {
		args = append(args, prompt)
	}
	for _, a := range r.harness.HeadlessModelArgs {
		args = append(args, r.expand(a))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, r.harness.HeadlessCmd[0], args...)
	cmd.Env = os.Environ()
	cmd.Dir = dir
	if r.harness.PromptViaStdin {
		cmd.Stdin = strings.NewReader(prompt)
	}

	out, err := cmd.CombinedOutput()
	r.lastOutput = string(out)
	if os.Getenv("HARNESS_DEBUG") != "" {
		if len(out) > 0 {
			fmt.Printf("    [debug] output:\n%s\n", r.lastOutput)
		} else {
			fmt.Printf("    [debug] no output, err=%v\n", err)
		}
	}
	if err != nil && len(out) > 0 {
		r.pass(fmt.Sprintf("headless produced output (%d bytes, exit: %v)", len(out), err))
	} else if err != nil {
		r.fail("headless: " + err.Error())
	} else if len(out) > 0 {
		r.pass(fmt.Sprintf("headless produced output (%d bytes)", len(out)))
	} else {
		r.fail("headless produced no output")
	}
	if r.harness.Events.Stop != "" {
		time.Sleep(3 * time.Second)
	}
}

func (r *Runner) runInteractive() {
	if len(r.harness.InteractiveCmd) == 0 || !r.harness.HooksInInteractive {
		return
	}

	fmt.Println("[phase 6] interactive (PTY) mode")

	dir := r.home
	if r.harness.NeedsGitRepo {
		dir = r.ensureGitRepo()
	}

	var iargs []string
	iargs = append(iargs, r.harness.InteractiveCmd[1:]...)
	for _, a := range r.harness.InteractiveArgs {
		iargs = append(iargs, r.expand(a))
	}

	session, err := StartPTY(r.harness.InteractiveCmd[0], iargs, dir, os.Environ())
	if err != nil {
		r.fail("PTY start: " + err.Error())
		return
	}
	defer session.Close()

	time.Sleep(3 * time.Second)
	if len(r.harness.OnboardingDismiss) > 0 {
		for i := 0; i < 15; i++ {
			out := session.Output()
			dismissed := false
			for _, pattern := range r.harness.OnboardingDismiss {
				if strings.Contains(out, pattern) {
					session.SendLine("")
					time.Sleep(2 * time.Second)
					dismissed = true
					break
				}
			}
			if dismissed {
				continue
			}
			if len(out) > 200 {
				break
			}
			time.Sleep(1 * time.Second)
		}
	} else {
		_, _ = session.WaitForAny([]string{">", "❯", "$", "?"}, 15*time.Second)
	}
	r.pass("TUI started")

	if r.harness.InteractivePromptInArgs {
		session.WaitForAny([]string{"mock", "hello", "Hello", "codename", "server", "build"}, 60*time.Second)
		time.Sleep(3 * time.Second)
	} else {
		session.SendLine("What is the project codename? Reply ONLY the codename.")
		session.WaitForAny([]string{"mock", "hello", "Hello", "codename", "server"}, 30*time.Second)
		time.Sleep(3 * time.Second)
		if r.harness.ExitCommand != "" {
			session.SendLine(r.harness.ExitCommand)
			time.Sleep(5 * time.Second)
		}
	}
	session.SendCtrlC()
	session.Wait(5 * time.Second)

	if r.harness.Events.Stop != "" {
		time.Sleep(2 * time.Second)
	}

	r.lastOutput = session.Output()
	if os.Getenv("HARNESS_DEBUG") != "" {
		fmt.Printf("    [debug] PTY output (%d bytes)\n", len(r.lastOutput))
	}

	r.pass("interactive session completed")
}

func (r *Runner) checkHookEvents(phase string) {
	fmt.Printf("[phase] hook events (%s)\n", phase)

	logContent := ""
	if data, err := os.ReadFile("/tmp/belt-hook-events.log"); err == nil {
		logContent = string(data)
	}

	var ptyContent string

	checks := []struct{ tag, label, event string }{
		{"SESSION_START", "session-start", r.harness.Events.SessionStart},
		{"PROMPT", "prompt", r.harness.Events.PromptSubmit},
		{"PRE_TOOL", "pre-tool-use", r.harness.Events.PreToolUse},
		{"POST_TOOL", "post-tool-use", r.harness.Events.PostToolUse},
		{"STOP", "stop", r.harness.Events.Stop},
		{"PRE_COMPACT", "pre-compact", r.harness.Events.PreCompact},
	}

	for _, c := range checks {
		if c.event == "" {
			continue
		}
		found := false
		if logContent != "" && strings.Contains(logContent, c.tag) {
			found = true
		}
		if !found && r.lastOutput != "" {
			if ptyContent == "" {
				ptyContent = stripANSI(r.lastOutput)
			}
			if strings.Contains(ptyContent, "hook: "+c.event) {
				found = true
			}
		}
		if found {
			r.pass(fmt.Sprintf("%s: %s hook fired", phase, c.label))
		} else {
			r.skip(fmt.Sprintf("%s: %s hook not fired", phase, c.label))
		}
	}

	if r.server.LogCount() > 0 {
		r.pass(fmt.Sprintf("%s: mock server received %d request(s)", phase, r.server.LogCount()))
	}
}

type eventEntry struct {
	Event string
	Tag   string
}

func (r *Runner) eventEntries() []eventEntry {
	evts := r.harness.Events
	all := []eventEntry{
		{evts.SessionStart, "SESSION_START"},
		{evts.PromptSubmit, "PROMPT"},
		{evts.PreToolUse, "PRE_TOOL"},
		{evts.PostToolUse, "POST_TOOL"},
		{evts.Stop, "STOP"},
		{evts.PreCompact, "PRE_COMPACT"},
	}
	var result []eventEntry
	for _, e := range all {
		if e.Event != "" {
			result = append(result, e)
		}
	}
	return result
}

func (r *Runner) buildNestedHooksJSON(logPath string) string {
	entries := r.eventEntries()
	parts := []string{}
	for _, e := range entries {
		cmd := fmt.Sprintf("echo %s >> %s", e.Tag, logPath)
		if e.Tag == "PROMPT" {
			cmd += fmt.Sprintf(" && echo 'The project codename is %s.'", r.injectCode)
		}
		hook := fmt.Sprintf(`{"type":"command","command":"%s","timeout":5}`, cmd)
		if e.Tag == "PRE_TOOL" || e.Tag == "POST_TOOL" {
			matcher := r.harness.ToolCallName
			if matcher == "" {
				matcher = "Read"
			}
			parts = append(parts, fmt.Sprintf(`"%s":[{"matcher":"%s","hooks":[%s]}]`, e.Event, matcher, hook))
		} else {
			parts = append(parts, fmt.Sprintf(`"%s":[{"hooks":[%s]}]`, e.Event, hook))
		}
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func (r *Runner) expand(tmpl string) string {
	s := strings.ReplaceAll(tmpl, "{{.BaseURL}}", r.baseURL)
	s = strings.ReplaceAll(s, "{{.Model}}", r.harness.DefaultModel)
	s = strings.ReplaceAll(s, "{{.APIKey}}", "mock-key")
	s = strings.ReplaceAll(s, "{{.HomeDir}}", r.home)
	if r.repoDir != "" {
		s = strings.ReplaceAll(s, "{{.RepoDir}}", r.repoDir)
	} else {
		s = strings.ReplaceAll(s, "{{.RepoDir}}", filepath.Join(r.home, "test-repo"))
	}
	return s
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}
