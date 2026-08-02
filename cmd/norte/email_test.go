package main

import "testing"

func TestNormalizeEmail(t *testing.T) {
	got := normalizeEmail("  Admin@Example.COM ")
	if got != "admin@example.com" {
		t.Fatalf("got %q", got)
	}
}

func TestIsValidEmail(t *testing.T) {
	if !isValidEmail("user@example.com") {
		t.Fatal("expected valid")
	}
	if isValidEmail("not-an-email") {
		t.Fatal("expected invalid")
	}
	if isValidEmail("Name <user@example.com>") {
		t.Fatal("expected bare address only")
	}
}
