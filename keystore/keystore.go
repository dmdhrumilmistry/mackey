// Package keystore provides functions to securely store, retrieve, and manage
// key-value pairs in the macOS Keychain via the go-keyring library.
package keystore

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/zalando/go-keyring"
)

// ErrNotFound is returned when a key does not exist in the keychain.
var ErrNotFound = errors.New("keystore: key not found")

// manifestKey is a reserved key used to track all stored keys for a service.
const manifestKey = "__mackey_manifest__"

// Set stores a key-value pair in the macOS Keychain under the given service name.
func Set(service, key, value string) error {
	if service == "" {
		return fmt.Errorf("keystore: service name must not be empty")
	}
	if key == "" {
		return fmt.Errorf("keystore: key must not be empty")
	}
	if key == manifestKey {
		return fmt.Errorf("keystore: key name %q is reserved", manifestKey)
	}
	if err := keyring.Set(service, key, value); err != nil {
		return err
	}
	return updateManifest(service, key, true)
}

// Get retrieves the value for the given key from the macOS Keychain.
// Returns ErrNotFound when the key does not exist.
func Get(service, key string) (string, error) {
	if service == "" {
		return "", fmt.Errorf("keystore: service name must not be empty")
	}
	if key == "" {
		return "", fmt.Errorf("keystore: key must not be empty")
	}
	val, err := keyring.Get(service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", ErrNotFound
	}
	return val, err
}

// Delete removes the key-value pair from the macOS Keychain.
// Returns ErrNotFound when the key does not exist.
func Delete(service, key string) error {
	if service == "" {
		return fmt.Errorf("keystore: service name must not be empty")
	}
	if key == "" {
		return fmt.Errorf("keystore: key must not be empty")
	}
	err := keyring.Delete(service, key)
	if errors.Is(err, keyring.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return updateManifest(service, key, false)
}

// List returns all keys that have been stored for the given service.
func List(service string) ([]string, error) {
	if service == "" {
		return nil, fmt.Errorf("keystore: service name must not be empty")
	}
	return readManifest(service)
}

// readManifest retrieves the list of keys from the manifest entry.
func readManifest(service string) ([]string, error) {
	raw, err := keyring.Get(service, manifestKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return []string{}, nil
	}
	if err != nil {
		return nil, err
	}
	var keys []string
	if err := json.Unmarshal([]byte(raw), &keys); err != nil {
		return nil, fmt.Errorf("keystore: corrupt manifest: %w", err)
	}
	return keys, nil
}

// updateManifest adds or removes a key from the manifest stored in the keychain.
func updateManifest(service, key string, add bool) error {
	keys, err := readManifest(service)
	if err != nil {
		return err
	}

	if add {
		// Avoid duplicate entries.
		for _, k := range keys {
			if k == key {
				return saveManifest(service, keys)
			}
		}
		keys = append(keys, key)
	} else {
		filtered := keys[:0]
		for _, k := range keys {
			if k != key {
				filtered = append(filtered, k)
			}
		}
		keys = filtered
	}

	return saveManifest(service, keys)
}

// saveManifest persists the key list to the keychain manifest entry.
func saveManifest(service string, keys []string) error {
	data, err := json.Marshal(keys)
	if err != nil {
		return fmt.Errorf("keystore: could not encode manifest: %w", err)
	}
	return keyring.Set(service, manifestKey, string(data))
}
