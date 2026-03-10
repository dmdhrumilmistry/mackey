package cmd

import (
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
