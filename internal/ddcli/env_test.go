package ddcli

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAllowSubmit(t *testing.T) {
	t.Setenv(EnvAllowSubmit, "")
	if AllowSubmit() {
		t.Fatal("expected false when unset")
	}

	t.Setenv(EnvAllowSubmit, "false")
	if AllowSubmit() {
		t.Fatal("expected false for false")
	}

	t.Setenv(EnvAllowSubmit, "true")
	if !AllowSubmit() {
		t.Fatal("expected true for true")
	}
}

func TestLoadDotEnv(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	contents := "# comment\nWIP_ALLOW_SUBMIT=true\nDD_CLI_BIN=/tmp/dd-cli\n"
	if err := os.WriteFile(envPath, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvAllowSubmit, "")
	t.Setenv("DD_CLI_BIN", "")
	// Clear so LoadDotEnv can set them (LookupEnv treats empty as set).
	_ = os.Unsetenv(EnvAllowSubmit)
	_ = os.Unsetenv("DD_CLI_BIN")

	if err := LoadDotEnv(envPath); err != nil {
		t.Fatal(err)
	}
	if !AllowSubmit() {
		t.Fatalf("AllowSubmit after load: %q", os.Getenv(EnvAllowSubmit))
	}
	if got := os.Getenv("DD_CLI_BIN"); got != "/tmp/dd-cli" {
		t.Fatalf("DD_CLI_BIN: %q", got)
	}
}

func TestLoadDotEnvDoesNotOverrideShell(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("WIP_ALLOW_SUBMIT=true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv(EnvAllowSubmit, "false")
	if err := LoadDotEnv(envPath); err != nil {
		t.Fatal(err)
	}
	if AllowSubmit() {
		t.Fatal("shell false should win over file true")
	}
}
