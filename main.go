package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	date, err := parseDate(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid date: %v\nUsage: gtd [dd/mm/yyyy]\n", err)
		os.Exit(1)
	}

	store, err := NewStore()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	p := tea.NewProgram(newModel(store, date), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("See you later!")
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
