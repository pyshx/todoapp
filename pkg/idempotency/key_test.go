package idempotency

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryStore_SetGet(t *testing.T) {
	store := NewInMemoryStore(1 * time.Hour)
	ctx := context.Background()

	resp := &Response{
		StatusCode: 200,
		Body:       []byte(`{"id": "123"}`),
	}

	key := "test-key"
	if err := store.Set(ctx, key, resp); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	got, ok := store.Get(ctx, key)
	if !ok {
		t.Fatal("expected to find key")
	}

	if got.StatusCode != resp.StatusCode {
		t.Errorf("expected status %d, got %d", resp.StatusCode, got.StatusCode)
	}
	if string(got.Body) != string(resp.Body) {
		t.Errorf("expected body %s, got %s", resp.Body, got.Body)
	}
}

func TestInMemoryStore_TTLExpiry(t *testing.T) {
	store := NewInMemoryStore(50 * time.Millisecond)
	ctx := context.Background()

	resp := &Response{
		StatusCode: 200,
		Body:       []byte(`{}`),
	}

	key := "expiring-key"
	if err := store.Set(ctx, key, resp); err != nil {
		t.Fatalf("failed to set: %v", err)
	}

	// Should exist immediately
	if _, ok := store.Get(ctx, key); !ok {
		t.Error("expected key to exist immediately")
	}

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	if _, ok := store.Get(ctx, key); ok {
		t.Error("expected key to be expired")
	}
}

func TestInMemoryStore_Delete(t *testing.T) {
	store := NewInMemoryStore(1 * time.Hour)
	ctx := context.Background()

	resp := &Response{
		StatusCode: 200,
		Body:       []byte(`{}`),
	}

	key := "delete-key"
	store.Set(ctx, key, resp)

	if _, ok := store.Get(ctx, key); !ok {
		t.Error("expected key to exist before delete")
	}

	store.Delete(ctx, key)

	if _, ok := store.Get(ctx, key); ok {
		t.Error("expected key to be deleted")
	}
}

func TestInMemoryStore_NotFound(t *testing.T) {
	store := NewInMemoryStore(1 * time.Hour)
	ctx := context.Background()

	if _, ok := store.Get(ctx, "nonexistent"); ok {
		t.Error("expected key to not exist")
	}
}

func TestGenerateKey_Deterministic(t *testing.T) {
	userID := "user-123"
	method := "CreateTask"
	body := []byte(`{"title": "test"}`)

	key1 := GenerateKey(userID, method, body)
	key2 := GenerateKey(userID, method, body)

	if key1 != key2 {
		t.Errorf("expected deterministic keys, got %s and %s", key1, key2)
	}
}

func TestGenerateKey_Different(t *testing.T) {
	body := []byte(`{"title": "test"}`)

	key1 := GenerateKey("user-1", "CreateTask", body)
	key2 := GenerateKey("user-2", "CreateTask", body)

	if key1 == key2 {
		t.Error("expected different keys for different users")
	}
}
