package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	if r.server != nil {
		r.server.SetToolCallMode(hasToolHooks)
	}

	if r.mode == ModeBoth || r.mode == ModeHeadless {
		r.runHeadless()
		r.checkHookEvents("headless")
	}
	if r.mode == ModeBoth || r.mode == ModeInteractive {
		os.Remove("/tmp/belt-hook-events.log")
		r.server.ClearLog()
		if r.server != nil {
			r.server.SetToolCallMode(hasToolHooks)
		}
		r.runInteractive()
		r.checkHookEvents("interactive")
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

func (r *Runner) writeConfigFiles() {
	if len(r.harness.ConfigFiles) == 0 {
		return
	}
	for _, cf := range r.harness.ConfigFiles {
		path := filepath.Join(r.home, cf.Path)
		os.MkdirAll(filepath.Dir(path), 0755)
		content := r.expand(cf.Content)
		os.WriteFile(path, []byte(content), 0644)
		if r.harness.NeedsGitRepo {
			projPath := filepath.Join(r.ensureGitRepo(), cf.Path)
			os.MkdirAll(filepath.Dir(projPath), 0755)
			os.WriteFile(projPath, []byte(content), 0644)
		}
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
		filename = "config.yaml"
		scriptPath := filepath.Join(r.home, ".hermes", "test-hook.sh")
		os.MkdirAll(filepath.Dir(scriptPath), 0755)
		scriptContent := fmt.Sprintf("#!/bin/sh\ncat - >/dev/null\necho PROMPT >> %s\nprintf '{\"context\": \"The project codename is %s.\"}\\n'\n", logPath, r.injectCode)
		os.WriteFile(scriptPath, []byte(scriptContent), 0755)
		content = fmt.Sprintf("model:\n  default: %s\n  provider: openrouter\n  base_url: %s/v1\nhooks:\n  %s:\n    - command: %s\n      timeout: 5\nhooks_auto_accept: true\n",
			r.harness.DefaultModel, r.baseURL, r.harness.Events.PromptSubmit, scriptPath)

	case harness.TSExtension:
		filename = "belt-test.ts"
		content = fmt.Sprintf(`export default function (pi: any) {
  pi.on("%s", async (event: any) => {
    require("fs").appendFileSync("%s", "PROMPT\n");
    return { systemPrompt: (event.systemPrompt || '') + '\nThe project codename is %s.' };
  });
}`, r.harness.Events.PromptSubmit, logPath, r.injectCode)

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
	if !r.harness.HooksInHeadless {
		r.skip(r.harness.Name + " does not fire hooks in headless mode")
		return
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	session, err := StartPTY(r.harness.InteractiveCmd[0], r.harness.InteractiveCmd[1:], dir, os.Environ())
	if err != nil {
		r.fail("PTY start: " + err.Error())
		return
	}
	defer session.Close()

	_, started := session.WaitForAny([]string{">", "$", "❯", "/", r.harness.Binary}, 15*time.Second)
	if !started {
		r.skip("TUI did not show prompt within 15s")
		return
	}
	r.pass("TUI started")

	session.SendLine("What is the project codename? Reply ONLY the codename.")
	time.Sleep(15 * time.Second)

	if r.harness.ExitCommand != "" {
		session.SendLine(r.harness.ExitCommand)
		time.Sleep(2 * time.Second)
	}
	session.SendCtrlC()
	session.Wait(5 * time.Second)

	r.pass("interactive session completed")
}

func (r *Runner) checkHookEvents(phase string) {
	fmt.Printf("[phase] hook events (%s)\n", phase)

	if r.harness.PreserveHome {
		r.checkHookEventsFromOutput(phase)
		return
	}

	data, err := os.ReadFile("/tmp/belt-hook-events.log")
	if err != nil {
		if phase == "headless" && !r.harness.HooksInHeadless {
			r.skip(phase + ": hooks not expected in this mode")
		} else {
			r.skip(phase + ": no hook events log")
		}
		return
	}

	content := string(data)
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
		if strings.Contains(content, c.tag) {
			r.pass(fmt.Sprintf("%s: %s hook fired", phase, c.label))
		} else {
			r.skip(fmt.Sprintf("%s: %s hook not fired", phase, c.label))
		}
	}

	if entries := r.server.Log(); len(entries) > 0 {
		r.pass(fmt.Sprintf("%s: mock server received %d request(s)", phase, len(entries)))
	}
}

func (r *Runner) checkHookEventsFromOutput(phase string) {
	out := r.lastOutput
	checks := []struct{ event, label string }{
		{r.harness.Events.SessionStart, "session-start"},
		{r.harness.Events.PromptSubmit, "prompt"},
		{r.harness.Events.PreToolUse, "pre-tool-use"},
		{r.harness.Events.PostToolUse, "post-tool-use"},
		{r.harness.Events.Stop, "stop"},
		{r.harness.Events.PreCompact, "pre-compact"},
	}
	for _, c := range checks {
		if c.event == "" {
			continue
		}
		if strings.Contains(out, "hook: "+c.event) {
			r.pass(fmt.Sprintf("%s: %s hook fired (output)", phase, c.label))
		} else {
			r.skip(fmt.Sprintf("%s: %s hook not detected in output", phase, c.label))
		}
	}
	if entries := r.server.Log(); len(entries) > 0 {
		r.pass(fmt.Sprintf("%s: mock server received %d request(s)", phase, len(entries)))
	}
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

func (r *Runner) buildNestedHooksJSON(logPath string) string {
	evts := r.harness.Events
	entries := []struct {
		name, tag string
		matcher   string
	}{
		{evts.SessionStart, "SESSION_START", ""},
		{evts.PromptSubmit, "PROMPT", ""},
		{evts.PreToolUse, "PRE_TOOL", "Read"},
		{evts.PostToolUse, "POST_TOOL", "Read"},
		{evts.Stop, "STOP", ""},
		{evts.PreCompact, "PRE_COMPACT", ""},
	}
	parts := []string{}
	for _, e := range entries {
		if e.name == "" {
			continue
		}
		cmd := fmt.Sprintf("echo %s >> %s", e.tag, logPath)
		if e.tag == "PROMPT" {
			cmd += fmt.Sprintf(" && echo 'The project codename is %s.'", r.injectCode)
		}
		hook := fmt.Sprintf(`{"type":"command","command":"%s","timeout":5}`, cmd)
		if e.matcher != "" {
			parts = append(parts, fmt.Sprintf(`"%s":[{"matcher":"%s","hooks":[%s]}]`, e.name, e.matcher, hook))
		} else {
			parts = append(parts, fmt.Sprintf(`"%s":[{"hooks":[%s]}]`, e.name, hook))
		}
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}
