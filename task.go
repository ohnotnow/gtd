package main

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Priority represents task urgency (A=highest, D=lowest).
type Priority string

const (
	PriorityA Priority = "A"
	PriorityB Priority = "B"
	PriorityC Priority = "C"
	PriorityD Priority = "D"
)

func (p Priority) Label() string {
	switch p {
	case PriorityA:
		return "A - Must do"
	case PriorityB:
		return "B - Should do"
	case PriorityC:
		return "C - Nice to do"
	case PriorityD:
		return "D - Delegate/defer"
	default:
		return string(p)
	}
}

func (p Priority) Color() lipgloss.Color {
	switch p {
	case PriorityA:
		return lipgloss.Color("#ef4444") // red
	case PriorityB:
		return lipgloss.Color("#f97316") // orange
	case PriorityC:
		return lipgloss.Color("#0ea5e9") // sky
	case PriorityD:
		return lipgloss.Color("#a1a1aa") // zinc
	default:
		return lipgloss.Color("#a1a1aa")
	}
}

func PriorityOptions() []huh.Option[Priority] {
	return []huh.Option[Priority]{
		huh.NewOption(PriorityA.Label(), PriorityA),
		huh.NewOption(PriorityB.Label(), PriorityB),
		huh.NewOption(PriorityC.Label(), PriorityC),
		huh.NewOption(PriorityD.Label(), PriorityD),
	}
}

// Status represents task completion state.
type Status int

const (
	StatusTodo       Status = 0
	StatusDone       Status = 1
	StatusInProgress Status = 2
)

func (s Status) Symbol() string {
	switch s {
	case StatusInProgress:
		return "▶"
	case StatusDone:
		return "✓"
	default:
		return ""
	}
}

func (s Status) PrintLabel() string {
	switch s {
	case StatusInProgress:
		return "WIP"
	case StatusDone:
		return "Yes"
	default:
		return ""
	}
}

// Task represents a single to-do item for a specific day.
type Task struct {
	ID            int64
	Date          string // yyyy-mm-dd
	Description   string
	Priority      Priority
	TimeEstimate  string
	Status        Status
	CarriedFromID *int64
}

func (t Task) WasCarriedOver() bool {
	return t.CarriedFromID != nil
}

func (t Task) DisplayDescription() string {
	if t.WasCarriedOver() {
		return t.Description + " (carried over)"
	}
	return t.Description
}

func (t Task) DoneDisplay() string {
	return t.Status.PrintLabel()
}

func filterTasks(tasks []Task, predicate func(Task) bool) []Task {
	var result []Task
	for _, t := range tasks {
		if predicate(t) {
			result = append(result, t)
		}
	}
	return result
}
