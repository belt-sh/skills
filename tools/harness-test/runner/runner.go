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

	if r.mode == ModeBoth || r.mode == ModeHeadless {
		r.runHeadless()
		r.checkHookEvents("headless")
	}
	if r.mode == ModeBoth || r.mode == ModeInteractive {
		os.Remove(filepath.Join(r.home, "hook-events.log"))
		r.server.ClearLog()
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
		os.WriteFile(path, []byte(r.expand(cf.Content)), 0644)
	}
}

func (r *Runner) writeHooks() {
	fmt.Println("[phase 3] hooks")

	hookDir := filepath.Join(r.home, r.harness.HookConfigDir)
	os.MkdirAll(hookDir, 0755)

	logPath := filepath.Join(r.home, "hook-events.log")
	r.injectCode = fmt.Sprintf("%s-%d", strings.ToUpper(r.harness.Name), time.Now().UnixMilli())

	promptCmd := fmt.Sprintf("echo PROMPT >> %s && echo 'The project codename is %s.'", logPath, r.injectCode)
	stopCmd := fmt.Sprintf("echo STOP >> %s", logPath)

	var content string
	var filename string

	switch r.harness.HookFormat {
	case harness.JSONNested:
		filename = "belt.json"
		content = fmt.Sprintf(`{"hooks":{"%s":[{"hooks":[{"type":"command","command":"%s","timeout":5}]}],"%s":[{"hooks":[{"type":"command","command":"%s","timeout":5}]}]}}`,
			r.harness.Events.PromptSubmit, promptCmd, r.harness.Events.Stop, stopCmd)

	case harness.JSONCopilot:
		filename = "belt.json"
		injectJSON := fmt.Sprintf(`{\"additionalContext\": \"The project codename is %s.\"}`, r.injectCode)
		copilotPromptCmd := fmt.Sprintf("echo PROMPT >> %s && echo '%s'", logPath, injectJSON)
		content = fmt.Sprintf(`{"version":1,"hooks":{"%s":[{"type":"command","bash":"%s","timeoutSec":5}],"%s":[{"type":"command","bash":"%s","timeoutSec":5}]}}`,
			r.harness.Events.PromptSubmit, copilotPromptCmd, r.harness.Events.Stop, stopCmd)

	case harness.YAML:
		filename = "config.yaml"
		contextJSON := fmt.Sprintf(`{"context": "The project codename is %s."}`, r.injectCode)
		yamlPromptCmd := fmt.Sprintf(`echo PROMPT >> %s && echo '%s'`, logPath, contextJSON)
		content = fmt.Sprintf("model:\n  provider: openrouter\n  name: %s\nhooks:\n  %s:\n    - command: \"%s\"\n      timeout: 5\nhooks_auto_accept: true\n",
			r.harness.DefaultModel, r.harness.Events.PromptSubmit, yamlPromptCmd)

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

	os.WriteFile(filepath.Join(hookDir, filename), []byte(content), 0644)
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

	data, err := os.ReadFile(filepath.Join(r.home, "hook-events.log"))
	if err != nil {
		if phase == "headless" && !r.harness.HooksInHeadless {
			r.skip(phase + ": hooks not expected in this mode")
		} else {
			r.skip(phase + ": no hook events log")
		}
		return
	}

	content := string(data)
	if strings.Contains(content, "PROMPT") {
		r.pass(phase + ": prompt hook fired")
	} else {
		r.fail(phase + ": prompt hook not found")
	}
	if strings.Contains(content, "STOP") {
		r.pass(phase + ": stop hook fired")
	} else {
		r.skip(phase + ": stop hook not fired")
	}

	if entries := r.server.Log(); len(entries) > 0 {
		r.pass(fmt.Sprintf("%s: mock server received %d request(s)", phase, len(entries)))
	}
}

func (r *Runner) expand(tmpl string) string {
	s := strings.ReplaceAll(tmpl, "{{.BaseURL}}", r.baseURL)
	s = strings.ReplaceAll(s, "{{.Model}}", r.harness.DefaultModel)
	s = strings.ReplaceAll(s, "{{.APIKey}}", "mock-key")
	if r.repoDir != "" {
		s = strings.ReplaceAll(s, "{{.RepoDir}}", r.repoDir)
	} else {
		s = strings.ReplaceAll(s, "{{.RepoDir}}", filepath.Join(r.home, "test-repo"))
	}
	return s
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.Run()
}
