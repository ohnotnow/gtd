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

	if err := s.AddTask("2025-01-15", "Buy milk", PriorityB, "30m", "default"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddTask("2025-01-15", "Fix server", PriorityA, "2h", "default"); err != nil {
		t.Fatal(err)
	}
	if err := s.AddTask("2025-01-16", "Other day task", PriorityC, "1h", "default"); err != nil {
		t.Fatal(err)
	}

	tasks, err := s.GetTasksForDate("2025-01-15", "default")
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

	s.AddTask("2025-01-15", "Original", PriorityB, "1h", "default")
	tasks, _ := s.GetTasksForDate("2025-01-15", "default")
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

	s.AddTask("2025-01-15", "To delete", PriorityB, "1h", "default")
	tasks, _ := s.GetTasksForDate("2025-01-15", "default")
	id := tasks[0].ID

	if err := s.DeleteTask(id); err != nil {
		t.Fatal(err)
	}

	tasks, _ = s.GetTasksForDate("2025-01-15", "default")
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks after delete, got %d", len(tasks))
	}
}

func TestMarkCompleteAndIncomplete(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Task", PriorityB, "1h", "default")
	tasks, _ := s.GetTasksForDate("2025-01-15", "default")
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

	s.AddTask("2025-01-15", "Incomplete A", PriorityA, "1h", "default")
	s.AddTask("2025-01-15", "Incomplete B", PriorityB, "2h", "default")
	s.AddTask("2025-01-15", "Complete", PriorityC, "30m", "default")

	tasks, _ := s.GetTasksForDate("2025-01-15", "default")
	// Mark the third task as complete
	for _, task := range tasks {
		if task.Description == "Complete" {
			s.MarkComplete(task.ID)
		}
	}

	candidates, err := s.GetCarryOverCandidates("2025-01-15", "2025-01-16", "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
}

func TestCarryOverTasks(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Carry me", PriorityA, "1h", "default")
	s.AddTask("2025-01-15", "Carry me too", PriorityB, "2h", "default")

	tasks, _ := s.GetTasksForDate("2025-01-15", "default")

	if err := s.CarryOverTasks(tasks, "2025-01-16", "default"); err != nil {
		t.Fatal(err)
	}

	carried, _ := s.GetTasksForDate("2025-01-16", "default")
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

	s.AddTask("2025-01-15", "Task A", PriorityA, "1h", "default")
	s.AddTask("2025-01-15", "Task B", PriorityB, "2h", "default")

	tasks, _ := s.GetTasksForDate("2025-01-15", "default")

	// Carry over all tasks
	s.CarryOverTasks(tasks, "2025-01-16", "default")

	// Now get candidates again - should be empty since all have been carried
	candidates, err := s.GetCarryOverCandidates("2025-01-15", "2025-01-16", "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(candidates) != 0 {
		t.Errorf("expected 0 candidates after carry-over, got %d", len(candidates))
	}
}

func TestGetLatestDateWithIncompleteTasks(t *testing.T) {
	s := newTestStore(t)

	// No tasks at all — should return empty string
	date, err := s.GetLatestDateWithIncompleteTasks("2025-01-20", "default")
	if err != nil {
		t.Fatal(err)
	}
	if date != "" {
		t.Errorf("expected empty string, got %q", date)
	}

	// Add tasks on two different days
	s.AddTask("2025-01-10", "Old task", PriorityA, "1h", "default")
	s.AddTask("2025-01-15", "Recent task", PriorityB, "2h", "default")

	// Looking before the 20th — should find the 15th
	date, err = s.GetLatestDateWithIncompleteTasks("2025-01-20", "default")
	if err != nil {
		t.Fatal(err)
	}
	if date != "2025-01-15" {
		t.Errorf("expected 2025-01-15, got %q", date)
	}

	// Mark the 15th task as complete — should now find the 10th
	tasks, _ := s.GetTasksForDate("2025-01-15", "default")
	s.MarkComplete(tasks[0].ID)

	date, err = s.GetLatestDateWithIncompleteTasks("2025-01-20", "default")
	if err != nil {
		t.Fatal(err)
	}
	if date != "2025-01-10" {
		t.Errorf("expected 2025-01-10, got %q", date)
	}

	// Looking before the 10th — no earlier tasks
	date, err = s.GetLatestDateWithIncompleteTasks("2025-01-10", "default")
	if err != nil {
		t.Fatal(err)
	}
	if date != "" {
		t.Errorf("expected empty string, got %q", date)
	}
}

func TestCopyIncompleteTasks(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Task A", PriorityA, "1h", "default")
	s.AddTask("2025-01-15", "Task B", PriorityB, "2h", "default")
	s.AddTask("2025-01-15", "Task C", PriorityC, "30m", "default")

	// Mark one as complete
	tasks, _ := s.GetTasksForDate("2025-01-15", "default")
	for _, task := range tasks {
		if task.Description == "Task B" {
			s.MarkComplete(task.ID)
		}
	}

	// Copy to a new date — should only get the 2 incomplete tasks
	if err := s.CopyIncompleteTasks("2025-01-15", "2025-01-20", "default"); err != nil {
		t.Fatal(err)
	}

	copied, _ := s.GetTasksForDate("2025-01-20", "default")
	if len(copied) != 2 {
		t.Fatalf("expected 2 copied tasks, got %d", len(copied))
	}

	for _, c := range copied {
		if c.Description == "Task B" {
			t.Error("completed task 'Task B' should not have been copied")
		}
		if c.IsCompleted {
			t.Errorf("copied task %q should not be completed", c.Description)
		}
	}
}

func TestGetTasksForEmptyDate(t *testing.T) {
	s := newTestStore(t)

	tasks, err := s.GetTasksForDate("2025-01-15", "default")
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestContextIsolation(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Work task", PriorityA, "2h", "work")
	s.AddTask("2025-01-15", "Personal task", PriorityB, "1h", "personal")

	workTasks, _ := s.GetTasksForDate("2025-01-15", "work")
	if len(workTasks) != 1 {
		t.Fatalf("expected 1 work task, got %d", len(workTasks))
	}
	if workTasks[0].Description != "Work task" {
		t.Errorf("expected 'Work task', got %q", workTasks[0].Description)
	}

	personalTasks, _ := s.GetTasksForDate("2025-01-15", "personal")
	if len(personalTasks) != 1 {
		t.Fatalf("expected 1 personal task, got %d", len(personalTasks))
	}
	if personalTasks[0].Description != "Personal task" {
		t.Errorf("expected 'Personal task', got %q", personalTasks[0].Description)
	}

	defaultTasks, _ := s.GetTasksForDate("2025-01-15", "default")
	if len(defaultTasks) != 0 {
		t.Errorf("expected 0 default tasks, got %d", len(defaultTasks))
	}
}

func TestCarryOverRespectsContext(t *testing.T) {
	s := newTestStore(t)

	s.AddTask("2025-01-15", "Work task", PriorityA, "2h", "work")
	s.AddTask("2025-01-15", "Personal task", PriorityB, "1h", "personal")

	// Only work tasks should be carry-over candidates for the work context
	workCandidates, _ := s.GetCarryOverCandidates("2025-01-15", "2025-01-16", "work")
	if len(workCandidates) != 1 {
		t.Fatalf("expected 1 work carry-over candidate, got %d", len(workCandidates))
	}

	// Carry over work tasks
	s.CarryOverTasks(workCandidates, "2025-01-16", "work")

	// The carried task should appear in work context on the new date
	carried, _ := s.GetTasksForDate("2025-01-16", "work")
	if len(carried) != 1 {
		t.Fatalf("expected 1 carried work task, got %d", len(carried))
	}

	// Personal context on the new date should still be empty
	personalCarried, _ := s.GetTasksForDate("2025-01-16", "personal")
	if len(personalCarried) != 0 {
		t.Errorf("expected 0 carried personal tasks, got %d", len(personalCarried))
	}
}
