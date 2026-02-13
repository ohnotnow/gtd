package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	date, err := parseDate(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid date: %v\n", err)
		fmt.Fprintf(os.Stderr, "Usage: gtd [dd/mm/yyyy]\n")
		os.Exit(1)
	}

	store, err := NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	for {
		tasks, err := store.GetTasksForDate(date)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load tasks: %v\n", err)
			os.Exit(1)
		}

		heading := formatHeading(date)
		DisplayTaskTable(tasks, heading)

		action, err := ShowMenu()
		if err != nil {
			break
		}

		if action == "quit" {
			PrintOutro("See you later!")
			break
		}

		if err := handleAction(store, &date, action); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
}

func parseDate(args []string) (string, error) {
	if len(args) < 2 {
		return time.Now().Format("2006-01-02"), nil
	}

	t, err := time.Parse("02/01/2006", args[1])
	if err != nil {
		return "", fmt.Errorf("expected dd/mm/yyyy, got %q", args[1])
	}
	return t.Format("2006-01-02"), nil
}

func formatHeading(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.Format("Monday 2 January 2006")
}

func tomorrow(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

func handleAction(store *Store, date *string, action string) error {
	switch action {
	case "add":
		return handleAdd(store, *date)
	case "done":
		return handleMarkDone(store, *date)
	case "undone":
		return handleMarkUndone(store, *date)
	case "edit":
		return handleEdit(store, *date)
	case "delete":
		return handleDelete(store, *date)
	case "carry":
		return handleCarryOver(store, *date)
	case "view":
		return handleViewAnotherDay(date)
	}
	return nil
}

func handleAdd(store *Store, date string) error {
	desc, priority, estimate, err := ShowAddTaskForm()
	if err != nil {
		return nil // user cancelled
	}

	if err := store.AddTask(date, desc, priority, estimate); err != nil {
		return err
	}

	PrintInfo("Task added.")
	return nil
}

func handleMarkDone(store *Store, date string) error {
	tasks, err := store.GetTasksForDate(date)
	if err != nil {
		return err
	}

	incomplete := filterTasks(tasks, func(t Task) bool { return !t.IsCompleted })
	if len(incomplete) == 0 {
		PrintWarning("No incomplete tasks.")
		return nil
	}

	taskID, err := SelectTask(incomplete, "Which task is done?")
	if err != nil {
		return nil
	}

	if err := store.MarkComplete(taskID); err != nil {
		return err
	}

	PrintInfo("Task marked as done.")
	return nil
}

func handleMarkUndone(store *Store, date string) error {
	tasks, err := store.GetTasksForDate(date)
	if err != nil {
		return err
	}

	completed := filterTasks(tasks, func(t Task) bool { return t.IsCompleted })
	if len(completed) == 0 {
		PrintWarning("No completed tasks.")
		return nil
	}

	taskID, err := SelectTask(completed, "Which task to mark as not done?")
	if err != nil {
		return nil
	}

	if err := store.MarkIncomplete(taskID); err != nil {
		return err
	}

	PrintInfo("Task marked as not done.")
	return nil
}

func handleEdit(store *Store, date string) error {
	tasks, err := store.GetTasksForDate(date)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		PrintWarning("No tasks to edit.")
		return nil
	}

	taskID, err := SelectTask(tasks, "Which task to edit?")
	if err != nil {
		return nil
	}

	task, err := store.GetTask(taskID)
	if err != nil {
		return err
	}

	desc, priority, estimate, err := ShowEditTaskForm(task)
	if err != nil {
		return nil
	}

	if err := store.UpdateTask(taskID, desc, priority, estimate); err != nil {
		return err
	}

	PrintInfo("Task updated.")
	return nil
}

func handleDelete(store *Store, date string) error {
	tasks, err := store.GetTasksForDate(date)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		PrintWarning("No tasks to delete.")
		return nil
	}

	taskID, err := SelectTask(tasks, "Which task to delete?")
	if err != nil {
		return nil
	}

	task, err := store.GetTask(taskID)
	if err != nil {
		return err
	}

	confirmed, err := ShowConfirm(fmt.Sprintf("Delete '%s'?", task.Description))
	if err != nil {
		return nil
	}

	if confirmed {
		if err := store.DeleteTask(taskID); err != nil {
			return err
		}
		PrintInfo("Task deleted.")
	}

	return nil
}

func handleCarryOver(store *Store, date string) error {
	toDate := tomorrow(date)

	candidates, err := store.GetCarryOverCandidates(date, toDate)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		tasks, err := store.GetTasksForDate(date)
		if err != nil {
			return err
		}
		incomplete := filterTasks(tasks, func(t Task) bool { return !t.IsCompleted })
		if len(incomplete) == 0 {
			PrintWarning("No incomplete tasks to carry over.")
		} else {
			PrintInfo("All incomplete tasks have already been carried over.")
		}
		return nil
	}

	t, _ := time.Parse("2006-01-02", toDate)
	PrintInfo(fmt.Sprintf("Carrying %d task(s) to %s:", len(candidates), t.Format("02/01/2006")))
	for _, task := range candidates {
		PrintInfo(fmt.Sprintf("  - %s", task.Description))
	}

	confirmed, err := ShowConfirm("Carry these over?")
	if err != nil {
		return nil
	}

	if confirmed {
		if err := store.CarryOverTasks(candidates, toDate); err != nil {
			return err
		}
		PrintInfo("Tasks carried over.")
	}

	return nil
}

func handleViewAnotherDay(date *string) error {
	input, err := ShowDateInput()
	if err != nil {
		return nil
	}

	t, err := time.Parse("02/01/2006", input)
	if err != nil {
		PrintWarning("Invalid date format. Please use dd/mm/yyyy.")
		return nil
	}

	*date = t.Format("2006-01-02")
	return nil
}
