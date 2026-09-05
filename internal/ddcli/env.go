package ddcli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// EnvAllowSubmitOrder is the env var that unlocks SubmitOrder (charges money).
// Set ALLOW_SUBMIT_ORDER=true in .env (gitignored) or the process environment.
const EnvAllowSubmitOrder = "ALLOW_SUBMIT_ORDER"

// LoadDotEnv reads KEY=VALUE lines from path into the process environment.
// Does not overwrite keys that are already set. Missing file is a no-op.
//
// Params:
//   - path (string) — usually ".env" in the repo root
//
// Returns:
//   - error — unreadable file (other than not-exist)
func LoadDotEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("ddcli: load dotenv: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		// Prefer an already-exported shell value over the file.
		if _, alreadySet := os.LookupEnv(key); alreadySet {
			continue
		}
		_ = os.Setenv(key, value)
	}
	return scanner.Err()
}

// AllowSubmit reports whether ALLOW_SUBMIT_ORDER is enabled (true/1/yes).
// Call LoadDotEnv(".env") first from demos so a local file is picked up.
func AllowSubmit() bool {
	rawValue := strings.TrimSpace(strings.ToLower(os.Getenv(EnvAllowSubmitOrder)))
	return rawValue == "true" || rawValue == "1" || rawValue == "yes"
}
