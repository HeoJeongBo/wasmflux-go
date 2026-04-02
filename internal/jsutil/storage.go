//go:build js && wasm

package jsutil

import (
	"encoding/json"
	"syscall/js"
)

// Storage wraps localStorage or sessionStorage with typed accessors.
type Storage struct {
	store js.Value
}

// LocalStorage returns a wrapper around window.localStorage.
func LocalStorage() *Storage {
	return &Storage{store: Global.Get("localStorage")}
}

// SessionStorage returns a wrapper around window.sessionStorage.
func SessionStorage() *Storage {
	return &Storage{store: Global.Get("sessionStorage")}
}

// Get retrieves a string value by key. Returns "" if not found.
func (s *Storage) Get(key string) string {
	v := s.store.Call("getItem", key)
	if v.IsNull() {
		return ""
	}
	return v.String()
}

// Set stores a string value by key.
func (s *Storage) Set(key, value string) {
	s.store.Call("setItem", key, value)
}

// GetJSON retrieves a value and unmarshals it from JSON into dst.
func (s *Storage) GetJSON(key string, dst any) error {
	v := s.store.Call("getItem", key)
	if v.IsNull() {
		return nil
	}
	return json.Unmarshal([]byte(v.String()), dst)
}

// SetJSON marshals a value to JSON and stores it.
func (s *Storage) SetJSON(key string, v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	s.store.Call("setItem", key, string(b))
	return nil
}

// Remove deletes a key from storage.
func (s *Storage) Remove(key string) {
	s.store.Call("removeItem", key)
}

// Clear removes all keys from storage.
func (s *Storage) Clear() {
	s.store.Call("clear")
}

// Len returns the number of items in storage.
func (s *Storage) Len() int {
	return s.store.Get("length").Int()
}

// Key returns the key at the given index.
func (s *Storage) Key(index int) string {
	v := s.store.Call("key", index)
	if v.IsNull() {
		return ""
	}
	return v.String()
}

// Has reports whether a key exists in storage.
func (s *Storage) Has(key string) bool {
	return !s.store.Call("getItem", key).IsNull()
}
