package main

import (
	"flag"
	"testing"
)

func TestParseFlagsTrailingDetected(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("dir", ".", "test flag")

	err := parseFlags(fs, []string{"positional", "--dir", "/tmp"})
	if err == nil {
		t.Fatal("expected error for trailing flag, got nil")
	}
}

func TestParseFlagsNormalOrder(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("dir", ".", "test flag")

	err := parseFlags(fs, []string{"--dir", "/tmp", "positional"})
	if err != nil {
		t.Fatalf("unexpected error for flags-before-args: %s", err)
	}
}

func TestParseFlagsNoArgs(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.String("dir", ".", "test flag")

	err := parseFlags(fs, []string{"--dir", "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestParseFlagsPositionalOnly(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)

	err := parseFlags(fs, []string{"pos1", "pos2"})
	if err != nil {
		t.Fatalf("unexpected error for positional-only args: %s", err)
	}
}
