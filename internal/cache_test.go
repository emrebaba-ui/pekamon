package internal

import (
	"bytes"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache := NewCache(5 * time.Second)

	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}

	if got, ok := cache.Get("missing"); ok || got.Val != nil {
		t.Fatalf("Get(missing) = (%v, %t), want (nil, false)", got, ok)
	}
}

func TestCacheAddGet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value []byte
	}{
		{name: "ordinary value", key: "pokemon", value: []byte("pikachu")},
		{name: "empty key", key: "", value: []byte("empty key")},
		{name: "empty value", key: "empty", value: []byte{}},
		{name: "nil value", key: "nil", value: nil},
		{name: "arbitrary bytes", key: "binary", value: []byte{0x00, 0xff, 0x10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewCache(5 * time.Second)
			cache.Add(tt.key, tt.value)

			got, ok := cache.Get(tt.key)
			if !ok {
				t.Fatal("Get() reported that an added key is absent")
			}
			if !bytes.Equal(got.Val, tt.value) {
				t.Errorf("Get() = %v, want %v", got, tt.value)
			}
			if tt.value != nil && got.Val == nil {
				t.Error("Get() returned nil for a non-nil value")
			}
		})
	}
}

func TestCacheGetMissingKey(t *testing.T) {
	cache := NewCache(5 * time.Second)
	cache.Add("present", []byte("value"))

	got, ok := cache.Get("missing")
	if ok {
		t.Fatal("Get(missing) reported that the key is present")
	}
	if got.Val != nil {
		t.Errorf("Get(missing) = %v, want nil", got.Val)
	}
}

func TestCacheAddOverwritesValue(t *testing.T) {
	cache := NewCache(5 * time.Second)
	cache.Add("pokemon", []byte("pikachu"))
	cache.Add("pokemon", []byte("eevee"))

	got, ok := cache.Get("pokemon")
	if !ok {
		t.Fatal("Get() reported that an overwritten key is absent")
	}
	if !bytes.Equal(got.Val, []byte("eevee")) {
		t.Errorf("Get() = %q, want %q", got.Val, []byte("eevee"))
	}
}

func TestCacheKeepsKeysIndependent(t *testing.T) {
	cache := NewCache(5 * time.Second)
	cache.Add("first", []byte("one"))
	cache.Add("second", []byte("two"))

	for key, want := range map[string][]byte{
		"first":  []byte("one"),
		"second": []byte("two"),
	} {
		got, ok := cache.Get(key)
		if !ok {
			t.Errorf("Get(%q) reported that the key is absent", key)
			continue
		}
		if !bytes.Equal(got.Val, want) {
			t.Errorf("Get(%q) = %q, want %q", key, got.Val, want)
		}
	}
}

func TestNewCacheCreatesIndependentCaches(t *testing.T) {
	first := NewCache(5 * time.Second)
	second := NewCache(5 * time.Second)

	first.Add("pokemon", []byte("pikachu"))

	if got, ok := second.Get("pokemon"); ok || got.Val != nil {
		t.Errorf("second cache Get(pokemon) = (%v, %t), want (nil, false)", got, ok)
	}
}

func TestReapLoop(t *testing.T) {
	c := NewCache(50 * time.Millisecond)

	c.Add("t1", []byte("t2"))
	if len(c.data) != 1 {
		t.Error("Reaploop test cannot be done, add method failed")
	}

	time.Sleep(51 * time.Millisecond)
	if len(c.data) != 0 {
		t.Errorf("reapLoop() does not function, left item: %d", len(c.data))
	}
}
