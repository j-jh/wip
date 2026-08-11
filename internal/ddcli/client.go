// Package ddcli wraps the DoorDash terminal CLI (`dd-cli`) as typed Go methods.
//
// Pattern for every command: build argv → run subprocess → decode JSON → check success.
// Callers should use methods like ListDeliveryAddresses, not RunCLICommand directly.
package ddcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const defaultCLIBinaryName = "dd-cli"

// CLIClient runs dd-cli as a subprocess with --json-output.
//
// Fields:
//   - BinaryPath (string) — path or name of the dd-cli binary.
//     Empty means: use DD_CLI_BIN if set, otherwise "dd-cli" on PATH.
type CLIClient struct {
	BinaryPath string
}

// resolveCLIBinaryPath picks which binary to exec: field → env → default name on PATH.
func (client *CLIClient) resolveCLIBinaryPath() string {
	if client != nil && client.BinaryPath != "" {
		return client.BinaryPath
	}
	if envPath := os.Getenv("DD_CLI_BIN"); envPath != "" {
		return envPath
	}
	return defaultCLIBinaryName
}

// BuildIntent formats the --intent string required by most dd-cli service commands.
//
// Params:
//   - summary (string) — who this is for and the goal (not a restatement of the command)
//   - userPrompt (string) — optional verbatim user prompt; empty omits that line
//
// Returns:
//   - string — formatted intent ready to pass as --intent
func BuildIntent(summary, userPrompt string) string {
	// strings.Builder is the usual Go way to build a string in pieces without
	// creating many temporary strings along the way.
	var intentBuilder strings.Builder
	intentBuilder.WriteString("Summary: ")
	intentBuilder.WriteString(summary)
	if userPrompt != "" {
		intentBuilder.WriteString("\nuser prompt/purpose: \"")
		intentBuilder.WriteString(userPrompt)
		intentBuilder.WriteString("\"")
	}
	return intentBuilder.String()
}

// RunCLICommand executes dd-cli with --json-output prepended to cliArgs.
//
// Params:
//   - ctx (context.Context) — cancel / timeout for the subprocess
//   - cliArgs (...string) — CLI args after the binary name (subcommand + flags)
//
// Returns:
//   - []byte — trimmed stdout (JSON when the command supports --json-output)
//   - error — non-zero exit; message includes stderr when present
//
// Notes: low-level helper. Prefer typed methods like ListDeliveryAddresses for callers.
func (client *CLIClient) RunCLICommand(ctx context.Context, cliArgs ...string) ([]byte, error) {
	// Always ask for JSON so every wrapper can decode the same envelope shape.
	commandArgs := make([]string, 0, len(cliArgs)+1)
	commandArgs = append(commandArgs, "--json-output")
	commandArgs = append(commandArgs, cliArgs...)

	// CommandContext stops the process if ctx is canceled or times out.
	cliCommand := exec.CommandContext(ctx, client.resolveCLIBinaryPath(), commandArgs...)

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	cliCommand.Stdout = &stdoutBuffer
	cliCommand.Stderr = &stderrBuffer

	runErr := cliCommand.Run()
	trimmedStdout := bytes.TrimSpace(stdoutBuffer.Bytes())
	if runErr != nil {
		errorMessage := strings.TrimSpace(stderrBuffer.String())
		if errorMessage == "" {
			errorMessage = strings.TrimSpace(stdoutBuffer.String())
		}
		if errorMessage == "" {
			errorMessage = runErr.Error()
		}
		// %w keeps the original exit error available to errors.Is / errors.As.
		return trimmedStdout, fmt.Errorf("ddcli: %w: %s", runErr, errorMessage)
	}
	return trimmedStdout, nil
}

// cliJSONOutputEnvelope is the outer --json-output wrapper from dd-cli.
// The real command payload is inside StructuredContent (another JSON object).
type cliJSONOutputEnvelope struct {
	// json.RawMessage delays parsing so we can unmarshal StructuredContent
	// into a different Go type per command.
	StructuredContent json.RawMessage `json:"structuredContent"`
	IsError           bool            `json:"isError"`
}

// decodeStructuredContent unmarshals dd-cli --json-output into commandResult from structuredContent.
//
// Params:
//   - cliStdout ([]byte) — full CLI stdout
//   - commandResult (any) — pointer to the command result struct
//
// Returns:
//   - error — bad JSON, missing structuredContent, or envelope isError
func decodeStructuredContent(cliStdout []byte, commandResult any) error {
	var outputEnvelope cliJSONOutputEnvelope
	if err := json.Unmarshal(cliStdout, &outputEnvelope); err != nil {
		return fmt.Errorf("ddcli: decode envelope: %w", err)
	}
	if outputEnvelope.IsError {
		return fmt.Errorf("ddcli: CLI reported isError")
	}
	if len(outputEnvelope.StructuredContent) == 0 {
		return fmt.Errorf("ddcli: missing structuredContent")
	}
	if err := json.Unmarshal(outputEnvelope.StructuredContent, commandResult); err != nil {
		return fmt.Errorf("ddcli: decode structuredContent: %w", err)
	}
	return nil
}
