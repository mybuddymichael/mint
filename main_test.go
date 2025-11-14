package main

import (
	"bytes"
	"context"
	"testing"
)

func TestCommandName(t *testing.T) {
	cmd := newCommand()
	if cmd.Name != "mt" {
		t.Errorf("expected command name 'mt', got '%s'", cmd.Name)
	}
}

func TestCommandRuns(t *testing.T) {
	cmd := newCommand()
	var buf bytes.Buffer
	cmd.Writer = &buf

	err := cmd.Run(context.Background(), []string{"mt"})
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

	err := cmd.Run(context.Background(), []string{"mt", "--help"})
	if err != nil {
		t.Errorf("cmd.Run() with --help returned error: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("expected help output, got none")
	}
}
