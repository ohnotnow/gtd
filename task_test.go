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
	done := Task{IsCompleted: true}
	if got := done.DoneDisplay(); got != "Yes" {
		t.Errorf("got %q, want %q", got, "Yes")
	}

	notDone := Task{IsCompleted: false}
	if got := notDone.DoneDisplay(); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
}

func TestFilterTasks(t *testing.T) {
	tasks := []Task{
		{Description: "a", IsCompleted: true},
		{Description: "b", IsCompleted: false},
		{Description: "c", IsCompleted: true},
	}

	completed := filterTasks(tasks, func(t Task) bool { return t.IsCompleted })
	if len(completed) != 2 {
		t.Errorf("expected 2 completed, got %d", len(completed))
	}

	incomplete := filterTasks(tasks, func(t Task) bool { return !t.IsCompleted })
	if len(incomplete) != 1 {
		t.Errorf("expected 1 incomplete, got %d", len(incomplete))
	}
}

func TestPriorityOptions(t *testing.T) {
	opts := PriorityOptions()
	if len(opts) != 4 {
		t.Fatalf("expected 4 options, got %d", len(opts))
	}
}
