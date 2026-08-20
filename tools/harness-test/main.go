package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/belt-sh/skills/tools/harness-test/harness"
	"github.com/belt-sh/skills/tools/harness-test/runner"
	"github.com/belt-sh/skills/tools/harness-test/server"
)

func main() {
	var (
		harnessName = flag.String("harness", "", "harness to test (or 'all')")
		mode        = flag.String("mode", "both", "test mode: headless, interactive, or both")
		listFlag    = flag.Bool("list", false, "list available harnesses")
		serverOnly  = flag.Bool("server", false, "run mock server only (no tests)")
	)
	flag.Parse()

	if *listFlag {
		fmt.Println("Available harnesses:")
		fmt.Printf("  %-12s %-10s %-8s %-10s\n", "NAME", "BINARY", "API", "HOOKS")
		names := make([]string, 0, len(harness.All))
		for name := range harness.All {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			h := harness.All[name]
			fmt.Printf("  %-12s %-10s %-8s %-10s\n",
				name, h.Binary, apiName(h.APIFormat), hookName(h.HookFormat))
		}
		return
	}

	srv := server.New()

	if *serverOnly {
		fmt.Println("Mock inference server")
		fmt.Println("Endpoints: /v1/chat/completions, /v1/responses, /v1/messages, /v1/models")
		fmt.Println("Utilities: GET /log, GET /log/count, DELETE /log, POST /response")
		// For server-only mode we need a different listener setup
		// For now, start with the random port
		baseURL, err := srv.Start()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Listening at %s\n", baseURL)
		select {} // block forever
	}

	if *harnessName == "" {
		fmt.Fprintln(os.Stderr, "usage: harness-test --harness <name|all>")
		fmt.Fprintln(os.Stderr, "       harness-test --list")
		fmt.Fprintln(os.Stderr, "       harness-test --server")
		os.Exit(1)
	}

	baseURL, err := srv.Start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start mock server: %v\n", err)
		os.Exit(1)
	}
	defer srv.Close()
	fmt.Printf("Mock server at %s\n\n", baseURL)

	var targets []string
	if *harnessName == "all" {
		for name := range harness.All {
			targets = append(targets, name)
		}
		sort.Strings(targets)
	} else {
		for _, name := range strings.Split(*harnessName, ",") {
			name = strings.TrimSpace(name)
			if _, ok := harness.All[name]; !ok {
				fmt.Fprintf(os.Stderr, "unknown harness: %s\n", name)
				os.Exit(1)
			}
			targets = append(targets, name)
		}
	}

	totalPassed, totalFailed, totalSkipped := 0, 0, 0
	var failed []string
	var results []runner.Result

	for _, name := range targets {
		h := harness.All[name]
		srv.ClearLog()
		r := runner.New(h, srv, baseURL)
		r.SetMode(*mode)
		result := r.Run()
		results = append(results, result)
		totalPassed += result.Passed
		totalFailed += result.Failed
		totalSkipped += result.Skipped
		if result.Failed > 0 {
			failed = append(failed, name)
		}
	}

	var totalDuration time.Duration
	fmt.Println("=== Summary ===")
	fmt.Printf("%-12s %6s %6s %6s %8s\n", "HARNESS", "PASS", "FAIL", "SKIP", "TIME")
	for _, r := range results {
		totalDuration += r.Duration
		fmt.Printf("%-12s %6d %6d %6d %8s\n", r.Harness, r.Passed, r.Failed, r.Skipped, r.Duration.Round(time.Second))
	}
	fmt.Printf("%-12s %6d %6d %6d %8s\n", "TOTAL", totalPassed, totalFailed, totalSkipped, totalDuration.Round(time.Second))
	if len(failed) > 0 {
		fmt.Printf("failures: %s\n", strings.Join(failed, ", "))
		os.Exit(1)
	}
}

func apiName(f harness.APIFormat) string {
	switch f {
	case harness.OpenAI:
		return "OpenAI"
	case harness.Responses:
		return "Resp"
	case harness.Anthropic:
		return "Anthro"
	case harness.Gemini:
		return "Gemini"
	default:
		return "?"
	}
}

func hookName(f harness.HookFormat) string {
	switch f {
	case harness.JSONNested:
		return "JSON"
	case harness.JSONFlat:
		return "JSON-flat"
	case harness.JSONCopilot:
		return "Copilot"
	case harness.TOML:
		return "TOML"
	case harness.YAML:
		return "YAML"
	case harness.TSExtension:
		return "TS-ext"
	case harness.TSPlugin:
		return "TS-plug"
	default:
		return "?"
	}
}
