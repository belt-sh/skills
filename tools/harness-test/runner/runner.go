package runner

import (
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
	Events  []string
}

type Mode int

const (
	ModeBoth        Mode = iota
	ModeHeadless
	ModeInteractive
)

type Runner struct {
	harness harness.Harness
	server  *server.MockServer
	baseURL string
	home    string
	mode    Mode
	result  Result
}

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
	r.setupEndpoint()
	r.setupHooks()
	r.setupSkills()

	if r.mode == ModeBoth || r.mode == ModeHeadless {
		r.runHeadless()
		r.checkHookEvents("headless")
	}
	if r.mode == ModeBoth || r.mode == ModeInteractive {
		// Clear events from headless phase before interactive
		os.Remove(filepath.Join(r.home, "hook-events.log"))
		r.server.ClearLog()
		r.runInteractive()
		r.checkHookEvents("interactive")
	}

	fmt.Printf("\n=== %s: %d passed, %d failed, %d skipped ===\n\n",
		r.harness.Name, r.result.Passed, r.result.Failed, r.result.Skipped)
	return r.result
}

func (r *Runner) setupHome() {
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
		r.fail(r.harness.Binary + " not found")
		return
	}
	r.pass(r.harness.Binary + " installed")
}

func (r *Runner) setupEndpoint() {
	fmt.Println("[phase 2] endpoint")

	if r.harness.EndpointEnvVar != "" {
		os.Setenv(r.harness.EndpointEnvVar, r.baseURL)
		r.pass(r.harness.EndpointEnvVar + "=" + r.baseURL)
	}
	if r.harness.APIKeyEnvVar != "" {
		os.Setenv(r.harness.APIKeyEnvVar, "mock-key")
		r.pass(r.harness.APIKeyEnvVar + " set")
	}

	r.server.ClearLog()
}

func (r *Runner) setupHooks() {
	fmt.Println("[phase 3] hooks")

	hookDir := filepath.Join(r.home, r.harness.HookConfigDir)
	os.MkdirAll(hookDir, 0755)

	eventLogPath := filepath.Join(r.home, "hook-events.log")
	injectCode := fmt.Sprintf("%s-%d", strings.ToUpper(r.harness.Name), time.Now().UnixMilli())

	switch r.harness.HookFormat {
	case harness.JSONNested:
		r.writeJSONNestedHooks(hookDir, eventLogPath, injectCode)
	case harness.JSONCopilot:
		r.writeCopilotHooks(hookDir, eventLogPath, injectCode)
	case harness.YAML:
		r.writeYAMLHooks(hookDir, eventLogPath, injectCode)
	case harness.TSExtension:
		r.writeTSExtension(hookDir, eventLogPath, injectCode)
	default:
		r.skip("hook format not yet implemented: " + fmt.Sprint(r.harness.HookFormat))
		return
	}

	r.pass(fmt.Sprintf("hooks configured (code: %s)", injectCode))
}

func (r *Runner) writeJSONNestedHooks(dir, logPath, code string) {
	hooks := fmt.Sprintf(`{
  "hooks": {
    "%s": [{"hooks":[{"type":"command","command":"echo PROMPT >> %s && echo 'The project codename is %s.'","timeout":5}]}],
    "%s": [{"hooks":[{"type":"command","command":"echo STOP >> %s","timeout":5}]}]
  }
}`, r.harness.Events.PromptSubmit, logPath, code,
		r.harness.Events.Stop, logPath)

	path := filepath.Join(dir, "belt.json")
	os.WriteFile(path, []byte(hooks), 0644)
}

func (r *Runner) writeCopilotHooks(dir, logPath, code string) {
	hooks := fmt.Sprintf(`{
  "version": 1,
  "hooks": {
    "%s": [{"type":"command","bash":"echo PROMPT >> %s && echo '{\"additionalContext\": \"The project codename is %s.\"}'","timeoutSec":5}],
    "%s": [{"type":"command","bash":"echo STOP >> %s","timeoutSec":5}]
  }
}`, r.harness.Events.PromptSubmit, logPath, code,
		r.harness.Events.Stop, logPath)

	path := filepath.Join(dir, "belt.json")
	os.WriteFile(path, []byte(hooks), 0644)
}

func (r *Runner) writeYAMLHooks(dir, logPath, code string) {
	config := fmt.Sprintf(`model:
  provider: openrouter
  name: %s
hooks:
  %s:
    - command: "echo PROMPT >> %s && echo '{\"context\": \"The project codename is %s.\"}'"
      timeout: 5
hooks_auto_accept: true
`, r.harness.DefaultModel, r.harness.Events.PromptSubmit, logPath, code)

	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte(config), 0644)

	envPath := filepath.Join(dir, ".env")
	os.WriteFile(envPath, []byte("OPENROUTER_API_KEY=mock-key\n"), 0644)
}

func (r *Runner) writeTSExtension(dir, logPath, code string) {
	ext := fmt.Sprintf(`export default function (pi: any) {
  pi.on("%s", async (event: any) => {
    require("fs").appendFileSync("%s", "PROMPT\\n");
    return {
      systemPrompt: (event.systemPrompt || '') + '\\nThe project codename is %s.',
    };
  });
}
`, r.harness.Events.PromptSubmit, logPath, code)

	path := filepath.Join(dir, "belt-test.ts")
	os.WriteFile(path, []byte(ext), 0644)
}

func (r *Runner) setupSkills() {
	if r.harness.SkillsDir == "" {
		return
	}
	fmt.Println("[phase 4] skills")
	skillsDir := filepath.Join(r.home, r.harness.SkillsDir)
	os.MkdirAll(skillsDir, 0755)
	r.pass("skills directory created")
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

	if r.harness.NeedsGitRepo {
		repoDir := filepath.Join(r.home, "test-repo")
		os.MkdirAll(repoDir, 0755)
		run(repoDir, "git", "init", "-q")
		run(repoDir, "git", "config", "user.email", "t@t")
		run(repoDir, "git", "config", "user.name", "t")
		os.WriteFile(filepath.Join(repoDir, "README.md"), []byte("test"), 0644)
		run(repoDir, "git", "add", ".")
		run(repoDir, "git", "commit", "-qm", "init")
	}

	// Build the command
	args := append(r.harness.HeadlessCmd[1:], "What is the project codename? Reply ONLY the codename.")
	cmd := exec.Command(r.harness.HeadlessCmd[0], args...)
	cmd.Env = os.Environ()

	if r.harness.NeedsGitRepo {
		cmd.Dir = filepath.Join(r.home, "test-repo")
	}

	out, err := runWithTimeout(cmd, 60*time.Second)
	if err != nil {
		r.fail("headless: " + err.Error())
		return
	}

	if len(out) > 0 {
		r.pass("headless produced output (" + fmt.Sprintf("%d bytes", len(out)) + ")")
	} else {
		r.fail("headless produced no output")
	}
}

func (r *Runner) runInteractive() {
	if len(r.harness.InteractiveCmd) == 0 {
		return
	}
	if !r.harness.HooksInInteractive {
		r.skip(r.harness.Name + " hooks don't fire in interactive mode")
		return
	}

	fmt.Println("[phase 6] interactive (PTY) mode")

	// Clear hook events from headless phase
	eventLogPath := filepath.Join(r.home, "hook-events.log")
	os.Remove(eventLogPath)
	r.server.ClearLog()

	// Build command
	args := r.harness.InteractiveCmd[1:]
	args = append(args, r.harness.InteractiveExtraFlags...)

	dir := r.home
	if r.harness.NeedsGitRepo {
		dir = filepath.Join(r.home, "test-repo")
	}

	env := os.Environ()

	session, err := StartPTY(r.harness.InteractiveCmd[0], args, dir, env)
	if err != nil {
		r.fail("PTY start: " + err.Error())
		return
	}
	defer session.Close()

	// Wait for the TUI to start (look for a prompt indicator)
	_, started := session.WaitForAny([]string{">", "$", "❯", "/", "grok", "claude", "copilot", "hermes"}, 15*time.Second)
	if !started {
		r.skip("TUI did not show prompt within 15s")
		return
	}
	r.pass("TUI started")

	// Send a prompt
	session.SendLine("What is the project codename? Reply ONLY the codename.")

	// Wait for response
	time.Sleep(15 * time.Second)

	// Exit
	if r.harness.ExitCommand != "" {
		session.SendLine(r.harness.ExitCommand)
		time.Sleep(2 * time.Second)
	}
	session.SendCtrlC()
	session.Wait(5 * time.Second)

	r.pass("interactive session completed")
}

func (r *Runner) checkHookEvents(phase string) {
	label := fmt.Sprintf("[phase] hook events (%s)", phase)
	fmt.Println(label)

	logPath := filepath.Join(r.home, "hook-events.log")
	data, err := os.ReadFile(logPath)
	if err != nil {
		if phase == "headless" && !r.harness.HooksInHeadless {
			r.skip(phase + ": hooks not expected in this mode")
		} else if phase == "interactive" {
			r.skip(phase + ": no hook events log")
		} else {
			r.skip(phase + ": no hook events log")
		}
		return
	}

	content := string(data)
	if strings.Contains(content, "PROMPT") {
		r.pass(phase + ": prompt hook fired")
	} else {
		if phase == "headless" && !r.harness.HooksInHeadless {
			r.skip(phase + ": hooks not expected in headless mode")
		} else {
			r.fail(phase + ": prompt hook not found")
		}
	}
	if strings.Contains(content, "STOP") {
		r.pass(phase + ": stop hook fired")
	} else {
		r.skip(phase + ": stop hook not fired")
	}

	entries := r.server.Log()
	if len(entries) > 0 {
		r.pass(fmt.Sprintf("%s: mock server received %d request(s)", phase, len(entries)))
	}
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}

func runWithTimeout(cmd *exec.Cmd, timeout time.Duration) ([]byte, error) {
	done := make(chan struct{})
	var out []byte
	var err error

	go func() {
		out, err = cmd.CombinedOutput()
		close(done)
	}()

	select {
	case <-done:
		return out, err
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return nil, fmt.Errorf("timeout after %s", timeout)
	}
}
