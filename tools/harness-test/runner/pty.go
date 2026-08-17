package runner

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"
)

// PTYSession wraps a command running inside a pseudoterminal.
type PTYSession struct {
	cmd    *exec.Cmd
	ptmx   *os.File
	output bytes.Buffer
	done   chan error
}

// StartPTY spawns a command in a PTY and begins reading output.
func StartPTY(name string, args []string, dir string, env []string) (*PTYSession, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = env

	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{Rows: 50, Cols: 200})
	if err != nil {
		return nil, fmt.Errorf("pty start: %w", err)
	}

	s := &PTYSession{
		cmd:  cmd,
		ptmx: ptmx,
		done: make(chan error, 1),
	}

	// Read output in background
	go func() {
		io.Copy(&s.output, ptmx)
		s.done <- cmd.Wait()
	}()

	return s, nil
}

// SendKeys writes raw bytes to the PTY (simulating keyboard input).
func (s *PTYSession) SendKeys(text string) {
	s.ptmx.WriteString(text)
}

// SendLine writes text followed by Enter.
func (s *PTYSession) SendLine(text string) {
	s.ptmx.WriteString(text + "\r")
}

// SendCtrlC sends Ctrl+C.
func (s *PTYSession) SendCtrlC() {
	s.ptmx.Write([]byte{0x03})
}

// WaitForText waits until the output buffer contains the target text,
// or until the timeout expires.
func (s *PTYSession) WaitForText(target string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if strings.Contains(s.output.String(), target) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// WaitForAny waits until the output contains any of the target strings.
func (s *PTYSession) WaitForAny(targets []string, timeout time.Duration) (string, bool) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		out := s.output.String()
		for _, t := range targets {
			if strings.Contains(out, t) {
				return t, true
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", false
}

// Output returns the current accumulated output (with ANSI codes).
func (s *PTYSession) Output() string {
	return s.output.String()
}

// CleanOutput returns output with ANSI escape codes stripped.
func (s *PTYSession) CleanOutput() string {
	return stripANSI(s.output.String())
}

// Close kills the process and closes the PTY.
func (s *PTYSession) Close() {
	if s.cmd.Process != nil {
		s.cmd.Process.Kill()
	}
	s.ptmx.Close()
}

// Wait waits for the process to exit with a timeout.
func (s *PTYSession) Wait(timeout time.Duration) error {
	select {
	case err := <-s.done:
		return err
	case <-time.After(timeout):
		s.Close()
		return fmt.Errorf("timeout after %s", timeout)
	}
}

// stripANSI removes ANSI escape sequences from text.
func stripANSI(s string) string {
	var out strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == 0x1b {
			// ESC sequence
			i++
			if i < len(s) && s[i] == '[' {
				// CSI sequence — skip until letter
				i++
				for i < len(s) && !((s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
					i++
				}
				if i < len(s) {
					i++
				}
			} else if i < len(s) && s[i] == ']' {
				// OSC sequence — skip until BEL or ST
				i++
				for i < len(s) && s[i] != 0x07 && !(i+1 < len(s) && s[i] == 0x1b && s[i+1] == '\\') {
					i++
				}
				if i < len(s) {
					i++
				}
			} else if i < len(s) {
				i++ // skip one char after ESC
			}
		} else if s[i] < 0x20 && s[i] != '\n' && s[i] != '\r' && s[i] != '\t' {
			i++ // skip other control chars
		} else {
			out.WriteByte(s[i])
			i++
		}
	}
	return out.String()
}
