package main

import "testing"

func TestRolloutHashDeterministic(t *testing.T) {
	first := RolloutHash("feature-x", "user-42")
	for i := 0; i < 1000; i++ {
		if got := RolloutHash("feature-x", "user-42"); got != first {
			t.Fatalf("RolloutHash not deterministic: %d != %d", got, first)
		}
	}
}

func TestRolloutHashDistinguishesInputs(t *testing.T) {
	a := RolloutHash("key", "user")
	b := RolloutHash("key", "other")
	if a == b {
		t.Fatal("expected different users to hash differently")
	}

	c := RolloutHash("keyA", "user")
	d := RolloutHash("keyB", "user")
	if c == d {
		t.Fatal("expected different keys to hash differently")
	}
}

func TestRolloutHashNullByteSeparator(t *testing.T) {
	// "ab" + "\x00" + "c" must differ from "a" + "\x00" + "bc".
	concatA := RolloutHash("ab", "c")
	concatB := RolloutHash("a", "bc")
	if concatA == concatB {
		t.Fatal("expected null-byte separation to prevent concatenation collisions")
	}
}
