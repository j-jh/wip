package ddcli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllowSubmit(t *testing.T) {
	t.Setenv(EnvAllowSubmitOrder, "")
	if AllowSubmit() {
		t.Fatal("expected false when unset")
	}

	t.Setenv(EnvAllowSubmitOrder, "false")
	if AllowSubmit() {
		t.Fatal("expected false for false")
	}

	t.Setenv(EnvAllowSubmitOrder, "true")
	if !AllowSubmit() {
		t.Fatal("expected true for true")
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "# comment\nALLOW_SUBMIT_ORDER=true\nDD_CLI_BIN=/tmp/dd-cli\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvAllowSubmitOrder, "")
	t.Setenv("DD_CLI_BIN", "")
	// Clear so LoadDotEnv can set them (LookupEnv treats empty as set).
	_ = os.Unsetenv(EnvAllowSubmitOrder)
	_ = os.Unsetenv("DD_CLI_BIN")

	if err := LoadDotEnv(envPath); err != nil {
		t.Fatal(err)
	}
	if !AllowSubmit() {
		t.Fatalf("AllowSubmit after load: %q", os.Getenv(EnvAllowSubmitOrder))
	}
	if got := os.Getenv("DD_CLI_BIN"); got != "/tmp/dd-cli" {
		t.Fatalf("DD_CLI_BIN: %q", got)
	}
}

func TestLoadDotEnvDoesNotOverrideShell(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("ALLOW_SUBMIT_ORDER=true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvAllowSubmitOrder, "false")
	if err := LoadDotEnv(envPath); err != nil {
		t.Fatal(err)
	}
	if AllowSubmit() {
		t.Fatal("shell false should win over file true")
	}
}
