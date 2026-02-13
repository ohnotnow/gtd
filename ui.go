package main

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7c3aed")).Padding(0, 1)
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#eab308"))
	outroStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7c3aed")).Bold(true)
)

func DisplayTaskTable(tasks []Task, heading string) {
	fmt.Println()
	fmt.Println(titleStyle.Render(heading))
	fmt.Println()

	if len(tasks) == 0 {
		PrintInfo("No tasks for this day.")
		return
	}

	rows := make([][]string, len(tasks))
	for i, t := range tasks {
		rows[i] = []string{
			strconv.Itoa(i + 1),
			t.DisplayDescription(),
			string(t.Priority),
			t.TimeEstimate,
			t.DoneDisplay(),
		}
	}

	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#555"))).
		Headers("#", "Task", "Priority", "Time", "Done").
		Rows(rows...).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().PaddingRight(1).PaddingLeft(1)
			if row == table.HeaderRow {
				return s.Bold(true)
			}
			// Colour the priority column based on value
			if col == 2 && row > 0 && row-1 < len(tasks) {
				return s.Foreground(tasks[row-1].Priority.Color())
			}
			return s
		})

	fmt.Println(t)

	completed := len(filterTasks(tasks, func(t Task) bool { return t.IsCompleted }))
	PrintInfo(fmt.Sprintf("%d/%d tasks completed", completed, len(tasks)))
}

func ShowMenu() (string, error) {
	var action string
	err := huh.NewSelect[string]().
		Title("What would you like to do?").
		Options(
			huh.NewOption("Add task", "add"),
			huh.NewOption("Mark task done", "done"),
			huh.NewOption("Mark task not done", "undone"),
			huh.NewOption("Edit task", "edit"),
			huh.NewOption("Delete task", "delete"),
			huh.NewOption("Carry over to tomorrow", "carry"),
			huh.NewOption("View another day", "view"),
			huh.NewOption("Quit", "quit"),
		).
		Value(&action).
		Run()

	return action, err
}

func ShowAddTaskForm() (description string, priority Priority, timeEstimate string, err error) {
	priority = PriorityB // default
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What do you need to do?").
				Value(&description).
				Validate(notEmpty("Description")),
			huh.NewSelect[Priority]().
				Title("Priority?").
				Options(PriorityOptions()...).
				Value(&priority),
			huh.NewInput().
				Title("Time estimate? (eg 30m, 2h, 1d)").
				Value(&timeEstimate).
				Validate(notEmpty("Time estimate")),
		),
	).Run()

	return
}

func ShowEditTaskForm(task Task) (description string, priority Priority, timeEstimate string, err error) {
	description = task.Description
	priority = task.Priority
	timeEstimate = task.TimeEstimate

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Description").
				Value(&description).
				Validate(notEmpty("Description")),
			huh.NewSelect[Priority]().
				Title("Priority?").
				Options(PriorityOptions()...).
				Value(&priority),
			huh.NewInput().
				Title("Time estimate?").
				Value(&timeEstimate).
				Validate(notEmpty("Time estimate")),
		),
	).Run()

	return
}

func SelectTask(tasks []Task, label string) (int64, error) {
	options := make([]huh.Option[int64], len(tasks))
	for i, t := range tasks {
		options[i] = huh.NewOption(t.Description, t.ID)
	}

	var taskID int64
	err := huh.NewSelect[int64]().
		Title(label).
		Options(options...).
		Value(&taskID).
		Run()

	return taskID, err
}

func ShowConfirm(message string) (bool, error) {
	var confirmed bool
	err := huh.NewConfirm().
		Title(message).
		Affirmative("Yes").
		Negative("No").
		Value(&confirmed).
		Run()

	return confirmed, err
}

func ShowDateInput() (string, error) {
	var date string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Date (dd/mm/yyyy)").
				Value(&date).
				Validate(notEmpty("Date")),
		),
	).Run()

	return date, err
}

func PrintInfo(msg string) {
	fmt.Println(infoStyle.Render(msg))
}

func PrintWarning(msg string) {
	fmt.Println(warningStyle.Render(msg))
}

func PrintOutro(msg string) {
	fmt.Println()
	fmt.Println(outroStyle.Render(msg))
}

func notEmpty(field string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}
