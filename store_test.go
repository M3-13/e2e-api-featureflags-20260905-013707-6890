package main

import (
	"fmt"
	"sync"
	"testing"
)

func TestStoreCreateAndGet(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "feature1", Enabled: true, Description: "d", RolloutPercent: 50}
	if err := s.Create(f); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	got, ok := s.Get("feature1")
	if !ok {
		t.Fatal("expected flag to exist")
	}
	if got.Key != "feature1" || !got.Enabled || got.Description != "d" || got.RolloutPercent != 50 {
		t.Fatalf("unexpected flag: %+v", got)
	}
}

func TestStoreCreateDuplicateReturnsError(t *testing.T) {
	s := NewStore()
	f := Flag{Key: "feature1"}
	if err := s.Create(f); err != nil {
		t.Fatalf("first create: %v", err)
	}
	if err := s.Create(f); err != ErrFlagExists {
		t.Fatalf("expected ErrFlagExists, got %v", err)
	}
}

func TestStoreList(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a"})
	_ = s.Create(Flag{Key: "b"})
	list := s.List()
	if len(list) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(list))
	}
}

func TestStoreListEmpty(t *testing.T) {
	s := NewStore()
	if list := s.List(); len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
}

func TestStoreUpdate(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a", Enabled: true})
	updated, ok := s.Update("a", Flag{Key: "a", Enabled: false, Description: "updated"})
	if !ok {
		t.Fatal("expected update to succeed")
	}
	if updated.Enabled {
		t.Fatal("expected enabled false")
	}
	if updated.Description != "updated" {
		t.Fatalf("expected updated description, got %q", updated.Description)
	}
	got, _ := s.Get("a")
	if got.Enabled {
		t.Fatal("expected stored flag enabled false")
	}
}

func TestStoreUpdateMissing(t *testing.T) {
	s := NewStore()
	if _, ok := s.Update("nope", Flag{Key: "nope"}); ok {
		t.Fatal("expected update to fail on missing key")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	_ = s.Create(Flag{Key: "a"})
	if !s.Delete("a") {
		t.Fatal("expected delete to succeed")
	}
	if _, ok := s.Get("a"); ok {
		t.Fatal("expected flag to be gone")
	}
	if s.Delete("a") {
		t.Fatal("expected second delete to fail")
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("flag-%d", n)
			_ = s.Create(Flag{Key: key})
			_, _ = s.Get(key)
			_ = s.List()
			s.Delete(key)
		}(i)
	}
	wg.Wait()
}
