package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseArgsDefaults(t *testing.T) {
	date, printMode, err := parseArgs([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if printMode {
		t.Error("expected printMode=false with no args")
	}
	today := time.Now().Format("2006-01-02")
	if date != today {
		t.Errorf("expected today %q, got %q", today, date)
	}
}

func TestParseArgsDateOnly(t *testing.T) {
	date, printMode, err := parseArgs([]string{"25/12/2025"})
	if err != nil {
		t.Fatal(err)
	}
	if printMode {
		t.Error("expected printMode=false")
	}
	if date != "2025-12-25" {
		t.Errorf("expected 2025-12-25, got %q", date)
	}
}

func TestParseArgsPrintOnly(t *testing.T) {
	date, printMode, err := parseArgs([]string{"--print"})
	if err != nil {
		t.Fatal(err)
	}
	if !printMode {
		t.Error("expected printMode=true")
	}
	today := time.Now().Format("2006-01-02")
	if date != today {
		t.Errorf("expected today %q, got %q", today, date)
	}
}

func TestParseArgsDateThenPrint(t *testing.T) {
	date, printMode, err := parseArgs([]string{"14/02/2026", "--print"})
	if err != nil {
		t.Fatal(err)
	}
	if !printMode {
		t.Error("expected printMode=true")
	}
	if date != "2026-02-14" {
		t.Errorf("expected 2026-02-14, got %q", date)
	}
}

func TestParseArgsPrintThenDate(t *testing.T) {
	date, printMode, err := parseArgs([]string{"--print", "14/02/2026"})
	if err != nil {
		t.Fatal(err)
	}
	if !printMode {
		t.Error("expected printMode=true")
	}
	if date != "2026-02-14" {
		t.Errorf("expected 2026-02-14, got %q", date)
	}
}

func TestParseArgsInvalidDate(t *testing.T) {
	_, _, err := parseArgs([]string{"not-a-date"})
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestPrintTasksEmpty(t *testing.T) {
	s := newTestStore(t)
	output := capturePrintTasks(t, s, "2025-06-01")

	if !strings.Contains(output, "No tasks for this day.") {
		t.Errorf("expected 'No tasks' message, got:\n%s", output)
	}
}

func TestPrintTasksWithTasks(t *testing.T) {
	s := newTestStore(t)
	s.AddTask("2025-06-01", "Fix server", PriorityA, "2h")
	s.AddTask("2025-06-01", "Buy milk", PriorityB, "30m")

	tasks, _ := s.GetTasksForDate("2025-06-01")
	s.MarkComplete(tasks[0].ID)

	output := capturePrintTasks(t, s, "2025-06-01")

	if !strings.Contains(output, "Fix server") {
		t.Errorf("expected 'Fix server' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Buy milk") {
		t.Errorf("expected 'Buy milk' in output, got:\n%s", output)
	}
	if !strings.Contains(output, "1/2 tasks completed") {
		t.Errorf("expected '1/2 tasks completed' in output, got:\n%s", output)
	}
}

func capturePrintTasks(t *testing.T, store *Store, date string) string {
	t.Helper()

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := printTasks(store, date)

	w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("printTasks returned error: %v", err)
	}

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
