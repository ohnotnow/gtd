package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	date, printMode, context, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\nUsage: gtd [dd/mm/yyyy] [--print] [--context <name>]\n", err)
		os.Exit(1)
	}

	store, err := NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	if printMode {
		if err := printTasks(store, date, context); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	p := tea.NewProgram(newModel(store, date, context), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("See you later!")
}

// parseArgs extracts the date, --print flag, and --context from command-line arguments.
// Flags and date can appear in any order.
func parseArgs(args []string) (date string, printMode bool, context string, err error) {
	date = time.Now().Format("2006-01-02")
	context = "default"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--print" {
			printMode = true
			continue
		}
		if arg == "--context" {
			if i+1 >= len(args) {
				return "", false, "", fmt.Errorf("--context requires a value")
			}
			i++
			context = args[i]
			continue
		}
		if strings.HasPrefix(arg, "--context=") {
			context = strings.TrimPrefix(arg, "--context=")
			if context == "" {
				return "", false, "", fmt.Errorf("--context requires a value")
			}
			continue
		}
		t, parseErr := time.Parse("02/01/2006", arg)
		if parseErr != nil {
			return "", false, "", fmt.Errorf("expected dd/mm/yyyy, got %q", arg)
		}
		date = t.Format("2006-01-02")
	}

	return date, printMode, context, nil
}

func printTasks(store *Store, date, context string) error {
	tasks, err := store.GetTasksForDate(date, context)
	if err != nil {
		return err
	}

	heading := formatHeading(date)
	fmt.Println(heading)
	fmt.Println()

	if len(tasks) == 0 {
		fmt.Println("No tasks for this day.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "#\tTask\tPriority\tTime\tDone")
	for i, t := range tasks {
		done := ""
		if t.IsCompleted {
			done = "Yes"
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n", i+1, t.DisplayDescription(), t.Priority, t.TimeEstimate, done)
	}
	w.Flush()

	completed := len(filterTasks(tasks, func(t Task) bool { return t.IsCompleted }))
	fmt.Printf("\n%d/%d tasks completed\n", completed, len(tasks))

	return nil
}
