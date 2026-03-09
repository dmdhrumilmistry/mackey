package keystore_test

import (
	"errors"
	"testing"

	"github.com/dmdhrumilmistry/mackey/keystore"
	"github.com/zalando/go-keyring"
)

func init() {
	// Use an in-memory mock so tests run without a real keychain (e.g. on CI).
	keyring.MockInit()
}

const testService = "mackey-test"

func TestSetAndGet(t *testing.T) {
	t.Run("set and get a value", func(t *testing.T) {
		err := keystore.Set(testService, "testkey", "testvalue")
		if err != nil {
			t.Fatalf("Set() error = %v", err)
		}

		val, err := keystore.Get(testService, "testkey")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if val != "testvalue" {
			t.Errorf("Get() = %q, want %q", val, "testvalue")
		}
	})
}

func TestDelete(t *testing.T) {
	t.Run("delete an existing key", func(t *testing.T) {
		_ = keystore.Set(testService, "deletekey", "val")

		err := keystore.Delete(testService, "deletekey")
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		_, err = keystore.Get(testService, "deletekey")
		if !errors.Is(err, keystore.ErrNotFound) {
			t.Errorf("Get() after delete expected ErrNotFound, got %v", err)
		}
	})

	t.Run("delete a non-existent key returns ErrNotFound", func(t *testing.T) {
		err := keystore.Delete(testService, "nosuchkey")
		if !errors.Is(err, keystore.ErrNotFound) {
			t.Errorf("Delete() expected ErrNotFound, got %v", err)
		}
	})
}

func TestGetNotFound(t *testing.T) {
	_, err := keystore.Get(testService, "nonexistentkey")
	if !errors.Is(err, keystore.ErrNotFound) {
		t.Errorf("Get() expected ErrNotFound, got %v", err)
	}
}

func TestList(t *testing.T) {
	svc := "mackey-test-list"
	keys := []string{"alpha", "beta", "gamma"}
	for _, k := range keys {
		if err := keystore.Set(svc, k, "value-"+k); err != nil {
			t.Fatalf("Set(%q) error = %v", k, err)
		}
	}

	listed, err := keystore.List(svc)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	listMap := make(map[string]bool, len(listed))
	for _, k := range listed {
		listMap[k] = true
	}
	for _, k := range keys {
		if !listMap[k] {
			t.Errorf("List() missing key %q", k)
		}
	}

	// After deletion the key should no longer appear.
	if err := keystore.Delete(svc, "beta"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	listed, err = keystore.List(svc)
	if err != nil {
		t.Fatalf("List() after delete error = %v", err)
	}
	for _, k := range listed {
		if k == "beta" {
			t.Errorf("List() still contains deleted key %q", k)
		}
	}
}

func TestValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		fn      func() error
		wantErr bool
	}{
		{
			name:    "Set empty service",
			fn:      func() error { return keystore.Set("", "k", "v") },
			wantErr: true,
		},
		{
			name:    "Set empty key",
			fn:      func() error { return keystore.Set("svc", "", "v") },
			wantErr: true,
		},
		{
			name:    "Get empty service",
			fn:      func() error { _, err := keystore.Get("", "k"); return err },
			wantErr: true,
		},
		{
			name:    "Get empty key",
			fn:      func() error { _, err := keystore.Get("svc", ""); return err },
			wantErr: true,
		},
		{
			name:    "Delete empty service",
			fn:      func() error { return keystore.Delete("", "k") },
			wantErr: true,
		},
		{
			name:    "Delete empty key",
			fn:      func() error { return keystore.Delete("svc", "") },
			wantErr: true,
		},
		{
			name:    "List empty service",
			fn:      func() error { _, err := keystore.List(""); return err },
			wantErr: true,
		},
		{
			name:    "Set key starting with digit",
			fn:      func() error { return keystore.Set("svc", "1INVALID", "v") },
			wantErr: true,
		},
		{
			name:    "Set key with hyphen",
			fn:      func() error { return keystore.Set("svc", "INVALID-KEY", "v") },
			wantErr: true,
		},
		{
			name:    "Set key with space",
			fn:      func() error { return keystore.Set("svc", "INVALID KEY", "v") },
			wantErr: true,
		},
		{
			name:    "Set valid key with underscore",
			fn:      func() error { return keystore.Set("svc-valid", "_VALID_KEY", "v") },
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if (err != nil) != tt.wantErr {
				t.Errorf("expected error=%v, got %v", tt.wantErr, err)
			}
		})
	}
}
