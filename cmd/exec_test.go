package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/zalando/go-keyring"
)

func init() {
	// Use an in-memory mock so tests run without a real keychain.
	keyring.MockInit()
}

// TestBuildExecEnvInjectsKeychain verifies that secrets from the keychain are
// present in the environment built for the child process.
func TestBuildExecEnvInjectsKeychain(t *testing.T) {
	const svc = "exec-test-inject"
	const key = "MACKEY_TEST_SECRET"
	const val = "hello-from-keychain"

	if err := keystore.Set(svc, key, val); err != nil {
		t.Fatalf("keystore.Set: %v", err)
	}
	t.Cleanup(func() { _ = keystore.Delete(svc, key) })

	env, err := buildExecEnv(svc, []string{"PATH=/usr/bin"})
	if err != nil {
		t.Fatalf("buildExecEnv: %v", err)
	}

	found := false
	for _, e := range env {
		if e == key+"="+val {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("env does not contain %s=%s; env=%v", key, val, env)
	}
}

// TestBuildExecEnvOverridesParent verifies that a keychain value replaces a
// variable that already exists in the parent environment.
func TestBuildExecEnvOverridesParent(t *testing.T) {
	const svc = "exec-test-override"
	const key = "MACKEY_OVERRIDE_VAR"
	const keychainVal = "from-keychain"
	const parentVal = "from-parent"

	if err := keystore.Set(svc, key, keychainVal); err != nil {
		t.Fatalf("keystore.Set: %v", err)
	}
	t.Cleanup(func() { _ = keystore.Delete(svc, key) })

	parentEnv := []string{key + "=" + parentVal, "PATH=/usr/bin"}
	env, err := buildExecEnv(svc, parentEnv)
	if err != nil {
		t.Fatalf("buildExecEnv: %v", err)
	}

	for _, e := range env {
		if strings.HasPrefix(e, key+"=") {
			if e != key+"="+keychainVal {
				t.Errorf("got %q, want %q", e, key+"="+keychainVal)
			}
			return
		}
	}
	t.Errorf("env missing %s; env=%v", key, env)
}

// TestBuildExecEnvNoDuplicates ensures the child env never contains two
// copies of the same variable when the parent env has the same key.
func TestBuildExecEnvNoDuplicates(t *testing.T) {
	const svc = "exec-test-dedup"
	const key = "MACKEY_DEDUP_VAR"

	if err := keystore.Set(svc, key, "keychain"); err != nil {
		t.Fatalf("keystore.Set: %v", err)
	}
	t.Cleanup(func() { _ = keystore.Delete(svc, key) })

	parentEnv := []string{key + "=parent", "HOME=/root"}
	env, err := buildExecEnv(svc, parentEnv)
	if err != nil {
		t.Fatalf("buildExecEnv: %v", err)
	}

	count := 0
	for _, e := range env {
		if strings.HasPrefix(e, key+"=") {
			count++
		}
	}
	if count != 1 {
		t.Errorf("env contains %d copies of %s, want exactly 1; env=%v", count, key, env)
	}
}

// TestBuildExecEnvPreservesParent verifies that unrelated parent env vars are
// passed through unchanged.
func TestBuildExecEnvPreservesParent(t *testing.T) {
	const svc = "exec-test-preserve"
	const key = "MACKEY_NEW_KEY"

	if err := keystore.Set(svc, key, "value"); err != nil {
		t.Fatalf("keystore.Set: %v", err)
	}
	t.Cleanup(func() { _ = keystore.Delete(svc, key) })

	parentEnv := []string{"PATH=/usr/bin", "HOME=/root", "USER=alice"}
	env, err := buildExecEnv(svc, parentEnv)
	if err != nil {
		t.Fatalf("buildExecEnv: %v", err)
	}

	for _, want := range parentEnv {
		found := false
		for _, e := range env {
			if e == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("parent var %q missing from child env; env=%v", want, env)
		}
	}
}

// TestExecChildProcessReceivesSecrets is an integration test that actually
// spawns a child process and verifies it receives the injected secrets.
// It uses printenv (POSIX) to read the value of the env var from the child's
// environment without any shell-expansion ambiguity.
func TestExecChildProcessReceivesSecrets(t *testing.T) {
	const svc = "exec-test-child"
	const key = "MACKEY_CHILD_SECRET"
	const val = "child-secret-value"

	if err := keystore.Set(svc, key, val); err != nil {
		t.Fatalf("keystore.Set: %v", err)
	}
	t.Cleanup(func() { _ = keystore.Delete(svc, key) })

	env, err := buildExecEnv(svc, []string{"PATH=/usr/bin:/bin"})
	if err != nil {
		t.Fatalf("buildExecEnv: %v", err)
	}

	// Use printenv to read the injected variable — no shell expansion involved.
	var out bytes.Buffer
	child := exec.Command("printenv", key) //nolint:gosec // key is a test constant, not user input
	child.Env = env
	child.Stdout = &out

	if err := child.Run(); err != nil {
		t.Fatalf("child process failed: %v", err)
	}

	got := strings.TrimRight(out.String(), "\n")
	if got != val {
		t.Errorf("child process got %q for %s, want %q", got, key, val)
	}
}

// TestExecSubprocessInheritsSecrets verifies that grandchild processes (spawned
// by the child) also inherit the injected environment, because environment
// variables are inherited across the entire process tree.
func TestExecSubprocessInheritsSecrets(t *testing.T) {
	const svc = "exec-test-grandchild"
	const key = "MACKEY_GRANDCHILD_SECRET"
	const val = "grandchild-secret-value"

	if err := keystore.Set(svc, key, val); err != nil {
		t.Fatalf("keystore.Set: %v", err)
	}
	t.Cleanup(func() { _ = keystore.Delete(svc, key) })

	env, err := buildExecEnv(svc, []string{"PATH=/usr/bin:/bin"})
	if err != nil {
		t.Fatalf("buildExecEnv: %v", err)
	}

	// Use sh -c 'printenv KEY' to simulate a subprocess (sh) spawning another
	// process (printenv). The secret must be available in the grandchild.
	var out bytes.Buffer
	child := exec.Command("sh", "-c", "printenv "+key) //nolint:gosec // key is a test constant, command injection not possible
	child.Env = env
	child.Stdout = &out

	if err := child.Run(); err != nil {
		t.Fatalf("child process failed: %v", err)
	}

	got := strings.TrimRight(out.String(), "\n")
	if got != val {
		t.Errorf("grandchild process got %q for %s, want %q", got, key, val)
	}
}
