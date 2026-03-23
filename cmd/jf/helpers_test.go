package main

import (
	"testing"

	"github.com/thebrokencube/files-with-a-dot/pkg/dendrik"
)

func TestParseTrailingFlagsHandled(t *testing.T) {
	// ff v4 handles interspersed flags natively, so trailing flags work
	fs := dendrik.NewFlagSet("test")
	fs.StringLong("dir", ".", "test flag")

	err := dendrik.Parse(fs, []string{"positional", "--dir", "/tmp"})
	if err != nil {
		t.Fatalf("ff v4 should handle interspersed flags, got: %s", err)
	}
}

func TestParseNormalOrder(t *testing.T) {
	fs := dendrik.NewFlagSet("test")
	dir := fs.StringLong("dir", ".", "test flag")

	err := dendrik.Parse(fs, []string{"--dir", "/tmp", "positional"})
	if err != nil {
		t.Fatalf("unexpected error for flags-before-args: %s", err)
	}
	if *dir != "/tmp" {
		t.Fatalf("dir: got %q, want %q", *dir, "/tmp")
	}
}

func TestParseNoArgs(t *testing.T) {
	fs := dendrik.NewFlagSet("test")
	dir := fs.StringLong("dir", ".", "test flag")

	err := dendrik.Parse(fs, []string{"--dir", "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if *dir != "/tmp" {
		t.Fatalf("dir: got %q, want %q", *dir, "/tmp")
	}
}

func TestParsePositionalOnly(t *testing.T) {
	fs := dendrik.NewFlagSet("test")

	err := dendrik.Parse(fs, []string{"pos1", "pos2"})
	if err != nil {
		t.Fatalf("unexpected error for positional-only args: %s", err)
	}
	args := fs.GetArgs()
	if len(args) != 2 {
		t.Fatalf("args: got %d, want 2", len(args))
	}
}

func TestParseUnknownFlagReturnsError(t *testing.T) {
	fs := dendrik.NewFlagSet("test")

	err := dendrik.Parse(fs, []string{"--bogus"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}
