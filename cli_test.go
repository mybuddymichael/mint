package main

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"
)

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

func TestCommandName(t *testing.T) {
	cmd := newCommand()
	if cmd.Name != "mint" {
		t.Errorf("expected command name 'mint', got '%s'", cmd.Name)
	}
}

func TestCommandRuns(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint"})
	if err != nil {
		t.Errorf("cmd.Run() returned error: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("expected output, got none")
	}
}

func TestCommandHelp(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "--help"})
	if err != nil {
		t.Errorf("cmd.Run() with --help returned error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected help output, got none")
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "--version"})
	if err != nil {
		t.Fatalf("--version flag failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, version) {
		t.Errorf("expected output to contain version '%s', got: %s", version, output)
	}
}

func TestVersionFlagShort(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mint", "-v"})
	if err != nil {
		t.Fatalf("-v flag failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, version) {
		t.Errorf("expected output to contain version '%s', got: %s", version, output)
	}
}

func TestShellCompletionEnabled(t *testing.T) {
	cmd := newCommand()
	if !cmd.EnableShellCompletion {
		t.Error("expected EnableShellCompletion to be true")
	}
}
