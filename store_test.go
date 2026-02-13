package main

import (
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStoreWithPath(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestAddAndGetTasks(t *testing.T) {
	s := newTestStore(t)

	if err := s.AddTask("2025-01-15", "Buy milk", PriorityB, "30m"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddTask("2025-01-15", "Fix server", PriorityA, "2h"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddTask("2025-01-16", "Other day task", PriorityC, "1h"); err != nil {
		t.Fatal(err)
	}

	tasks, err := s.GetTasksForDate("2025-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}

	// Should be ordered by priority: A before B
	if tasks[0].Description != "Fix server" {
		t.Errorf("expected first task 'Fix server', got %q", tasks[0].Description)
	}
	if tasks[1].Description != "Buy milk" {
		t.Errorf("expected second task 'Buy milk', got %q", tasks[1].Description)
	}
}

func TestUpdateTask(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Original", PriorityB, "1h")
	tasks, _ := s.GetTasksForDate("2025-01-15")
	id := tasks[0].ID

	if err := s.UpdateTask(id, "Updated", PriorityA, "2h"); err != nil {
		t.Fatal(err)
	}

	task, err := s.GetTask(id)
	if err != nil {
		t.Fatal(err)
	}
	if task.Description != "Updated" {
		t.Errorf("description = %q, want %q", task.Description, "Updated")
	}
	if task.Priority != PriorityA {
		t.Errorf("priority = %q, want %q", task.Priority, PriorityA)
	}
	if task.TimeEstimate != "2h" {
		t.Errorf("time_estimate = %q, want %q", task.TimeEstimate, "2h")
	}
}

func TestDeleteTask(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "To delete", PriorityB, "1h")
	tasks, _ := s.GetTasksForDate("2025-01-15")
	id := tasks[0].ID

	if err := s.DeleteTask(id); err != nil {
		t.Fatal(err)
	}

	tasks, _ = s.GetTasksForDate("2025-01-15")
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestMarkCompleteAndIncomplete(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Task", PriorityB, "1h")
	tasks, _ := s.GetTasksForDate("2025-01-15")
	id := tasks[0].ID

	if tasks[0].IsCompleted {
		t.Error("new task should not be completed")
	}

	s.MarkComplete(id)
	task, _ := s.GetTask(id)
	if !task.IsCompleted {
		t.Error("task should be completed after MarkComplete")
	}

	s.MarkIncomplete(id)
	task, _ = s.GetTask(id)
	if task.IsCompleted {
		t.Error("task should not be completed after MarkIncomplete")
	}
}

func TestCarryOverCandidates(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Incomplete A", PriorityA, "1h")
	s.AddTask("2025-01-15", "Incomplete B", PriorityB, "2h")
	s.AddTask("2025-01-15", "Complete", PriorityC, "30m")

	tasks, _ := s.GetTasksForDate("2025-01-15")
	// Mark the third task as complete
	for _, task := range tasks {
		if task.Description == "Complete" {
			s.MarkComplete(task.ID)
		}
	}

	candidates, err := s.GetCarryOverCandidates("2025-01-15", "2025-01-16")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
}

func TestCarryOverTasks(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Carry me", PriorityA, "1h")
	s.AddTask("2025-01-15", "Carry me too", PriorityB, "2h")

	tasks, _ := s.GetTasksForDate("2025-01-15")

	if err := s.CarryOverTasks(tasks, "2025-01-16"); err != nil {
		t.Fatal(err)
	}

	carried, _ := s.GetTasksForDate("2025-01-16")
	if len(carried) != 2 {
		t.Fatalf("expected 2 carried tasks, got %d", len(carried))
	}

	// Carried tasks should reference originals
	for _, c := range carried {
		if !c.WasCarriedOver() {
			t.Errorf("carried task %q should have CarriedFromID set", c.Description)
		}
		if c.IsCompleted {
			t.Errorf("carried task %q should not be completed", c.Description)
		}
	}
}

func TestCarryOverExcludesAlreadyCarried(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Task A", PriorityA, "1h")
	s.AddTask("2025-01-15", "Task B", PriorityB, "2h")

	tasks, _ := s.GetTasksForDate("2025-01-15")

	// Carry over all tasks
	s.CarryOverTasks(tasks, "2025-01-16")

	// Now get candidates again - should be empty since all have been carried
	candidates, err := s.GetCarryOverCandidates("2025-01-15", "2025-01-16")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates after carry-over, got %d", len(candidates))
	}
}

func TestGetTasksForEmptyDate(t *testing.T) {
	s := newTestStore(t)

	tasks, err := s.GetTasksForDate("2025-01-15")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}
