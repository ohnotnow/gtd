package main

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestPriorityLabel(t *testing.T) {
	tests := []struct {
		p    Priority
		want string
	}{
		{PriorityA, "A - Must do"},
		{PriorityB, "B - Should do"},
		{PriorityC, "C - Nice to do"},
		{PriorityD, "D - Delegate/defer"},
	}
	for _, tt := range tests {
		if got := tt.p.Label(); got != tt.want {
			t.Errorf("Priority(%s).Label() = %q, want %q", tt.p, got, tt.want)
		}
	}
}

func TestPriorityColor(t *testing.T) {
	tests := []struct {
		p    Priority
		want lipgloss.Color
	}{
		{PriorityA, lipgloss.Color("#ef4444")},
		{PriorityB, lipgloss.Color("#f97316")},
		{PriorityC, lipgloss.Color("#0ea5e9")},
		{PriorityD, lipgloss.Color("#a1a1aa")},
	}
	for _, tt := range tests {
		if got := tt.p.Color(); got != tt.want {
			t.Errorf("Priority(%s).Color() = %v, want %v", tt.p, got, tt.want)
		}
	}
}

func TestWasCarriedOver(t *testing.T) {
	task := Task{Description: "normal task"}
	if task.WasCarriedOver() {
		t.Error("task without CarriedFromID should not be carried over")
	}

	id := int64(42)
	carried := Task{Description: "carried task", CarriedFromID: &id}
	if !carried.WasCarriedOver() {
		t.Error("task with CarriedFromID should be carried over")
	}
}

func TestDisplayDescription(t *testing.T) {
	normal := Task{Description: "do laundry"}
	if got := normal.DisplayDescription(); got != "do laundry" {
		t.Errorf("got %q, want %q", got, "do laundry")
	}

	id := int64(1)
	carried := Task{Description: "do laundry", CarriedFromID: &id}
	if got := carried.DisplayDescription(); got != "do laundry (carried over)" {
		t.Errorf("got %q, want %q", got, "do laundry (carried over)")
	}
}

func TestDoneDisplay(t *testing.T) {
	done := Task{Status: StatusDone}
	if got := done.DoneDisplay(); got != "Yes" {
		t.Errorf("got %q, want %q", got, "Yes")
	}

	notDone := Task{Status: StatusTodo}
	if got := notDone.DoneDisplay(); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}

	wip := Task{Status: StatusInProgress}
	if got := wip.DoneDisplay(); got != "WIP" {
		t.Errorf("got %q, want %q", got, "WIP")
	}
}

func TestStatusSymbol(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusTodo, ""},
		{StatusInProgress, "▶"},
		{StatusDone, "✓"},
	}
	for _, tt := range tests {
		if got := tt.s.Symbol(); got != tt.want {
			t.Errorf("Status(%d).Symbol() = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestStatusPrintLabel(t *testing.T) {
	tests := []struct {
		s    Status
		want string
	}{
		{StatusTodo, ""},
		{StatusInProgress, "WIP"},
		{StatusDone, "Yes"},
	}
	for _, tt := range tests {
		if got := tt.s.PrintLabel(); got != tt.want {
			t.Errorf("Status(%d).PrintLabel() = %q, want %q", tt.s, got, tt.want)
		}
	}
}

func TestFilterTasks(t *testing.T) {
	tasks := []Task{
		{Description: "a", Status: StatusDone},
		{Description: "b", Status: StatusTodo},
		{Description: "c", Status: StatusDone},
		{Description: "d", Status: StatusInProgress},
	}

	completed := filterTasks(tasks, func(t Task) bool { return t.Status == StatusDone })
	if len(completed) != 2 {
		t.Errorf("expected 2 completed, got %d", len(completed))
	}

	incomplete := filterTasks(tasks, func(t Task) bool { return t.Status != StatusDone })
	if len(incomplete) != 2 {
		t.Errorf("expected 2 incomplete, got %d", len(incomplete))
	}
}

func TestPriorityOptions(t *testing.T) {
	opts := PriorityOptions()
	if len(opts) != 4 {
		t.Fatalf("expected 4 options, got %d", len(opts))
	}
}
