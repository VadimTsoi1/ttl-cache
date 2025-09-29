package cache

import (
	"testing"
	"time"
)

// До истечения TTL значение доступно
func TestSetGet_BeforeExpiry(t *testing.T) {
	c := NewCache(0)
	defer c.Close()

	c.Set("k", "v", 150*time.Millisecond)
	if got, ok := c.Get("k"); !ok || got != "v" {
		t.Fatalf("want v, true; got %q,%v", got, ok)
	}
}

// После истечения TTL запись удаляется на чтении.
func TestGet_ExpiresOnRead(t *testing.T) {
	c := NewCache(0)
	defer c.Close()

	c.Set("k", "v", 50*time.Millisecond)
	time.Sleep(90 * time.Millisecond)

	if got, ok := c.Get("k"); ok || got != "" {
		t.Fatalf("want \"\", false after expiry; got %q, %v", got, ok)
	}
}

// Фоновая очистка удаляет просроченные записи даже без чтения
func TestCleanup_RemovesExpired(t *testing.T) {
	c := NewCache(30 * time.Millisecond)
	defer c.Close()
	c.CleanUp()

	c.Set("a", "1", 40*time.Millisecond)
	c.Set("b", "2", 40*time.Millisecond)

	time.Sleep(120 * time.Millisecond)

	if _, ok := c.Get("a"); ok {
		t.Fatalf("key a should be expired and removed by cleanup")
	}
	if _, ok := c.Get("b"); ok {
		t.Fatal("key b shoulf be expired and removed by cleanup")
	}
}

// Бонус методы работают
func TestExistAndKeys(t *testing.T) {
	c := NewCache(0)
	defer c.Close()

	c.Set("x", "1", 100*time.Millisecond)
	if !c.Exists("x") {
		t.Fatalf("Exists should be true")
	}

	keys := c.Keys()
	if len(keys) == 0 {
		t.Fatalf("Keys should return at least one key")
	}
}
