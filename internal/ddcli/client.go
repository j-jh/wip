package ddcli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const defaultBin = "dd-cli"

// Client runs dd-cli as a subprocess with --json-output.
//
// Fields:
//   - Bin (string) — path or name of the dd-cli binary.
//     Empty means: use DD_CLI_BIN if set, otherwise "dd-cli" on PATH.
type Client struct {
	Bin string
}

func (c *Client) bin() string {
	if c != nil && c.Bin != "" {
		return c.Bin
	}
	if env := os.Getenv("DD_CLI_BIN"); env != "" {
		return env
	}
	return defaultBin
}

// Intent builds the --intent string required by most dd-cli service commands.
//
// Params:
//   - summary (string) — who this is for and the goal (not a restatement of the command)
//   - userPrompt (string) — optional verbatim user prompt; empty omits that line
//
// Returns:
//   - string — formatted intent ready to pass as --intent
func Intent(summary, userPrompt string) string {
	var b strings.Builder
	b.WriteString("Summary: ")
	b.WriteString(summary)
	if userPrompt != "" {
		b.WriteString("\nuser prompt/purpose: \"")
		b.WriteString(userPrompt)
		b.WriteString("\"")
	}
	return b.String()
}

// Run executes dd-cli with --json-output prepended to args.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the subprocess
//   - args (...string) — CLI args after the binary name (subcommand + flags)
//
// Returns:
//   - []byte — trimmed stdout (JSON when the command supports --json-output)
//   - error — non-zero exit; message includes stderr when present
//
// Notes: low-level helper. Prefer typed methods like AddressList for callers.
func (c *Client) Run(ctx context.Context, args ...string) ([]byte, error) {
	full := make([]string, 0, len(args)+1)
	full = append(full, "--json-output")
	full = append(full, args...)

	cmd := exec.CommandContext(ctx, c.bin(), full...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	out := bytes.TrimSpace(stdout.Bytes())
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = strings.TrimSpace(stdout.String())
		}
		if msg == "" {
			msg = err.Error()
		}
		return out, fmt.Errorf("ddcli: %w: %s", err, msg)
	}
	return out, nil
}
