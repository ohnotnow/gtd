package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7c3aed")).Padding(0, 1)
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#22c55e")).Italic(true)
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
)

// App modes
type mode int

const (
	modeTable mode = iota
	modeAdd
	modeEdit
	modeConfirmDelete
	modeConfirmCarry
	modeViewDate
	modeFilter
)

type model struct {
	store   *Store
	date    string
	context string
	tasks   []Task
	table   table.Model
	mode    mode
	form    *huh.Form
	status  string
	width   int
	height  int

	// Form field bindings (pointer receiver keeps addresses stable)
	formDesc     string
	formPriority Priority
	formEstimate string
	formDate     string
	formConfirm  bool

	// Filter
	filterText    string
	filteredTasks []Task

	// Context for current action
	editTaskID          int64
	carryCandidates     []Task
	latestDateWithTasks string
}

func newModel(store *Store, date, context string) *model {
	m := &model{
		store:   store,
		date:    date,
		context: context,
		width:   80,
	}
	m.refreshTasks()
	return m
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.rebuildTable()
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.mode {
	case modeTable:
		return m.updateTable(msg)
	case modeFilter:
		return m.updateFilter(msg)
	default:
		return m.updateForm(msg)
	}
}

func (m *model) View() string {
	var s strings.Builder

	s.WriteString("\n")
	heading := formatHeading(m.date)
	if m.context != "default" {
		heading += " · " + m.context
	}
	s.WriteString(titleStyle.Render(heading))
	s.WriteString("\n\n")

	switch m.mode {
	case modeTable, modeFilter:
		if len(m.tasks) == 0 {
			s.WriteString(infoStyle.Render("  No tasks for this day."))
			s.WriteString("\n")
			if m.latestDateWithTasks != "" {
				lt, _ := time.Parse("2006-01-02", m.latestDateWithTasks)
				s.WriteString("\n")
				s.WriteString(helpStyle.Render(fmt.Sprintf("  Press i to import tasks from %s", lt.Format("Monday 2 January 2006"))))
				s.WriteString("\n")
			}
		} else {
			s.WriteString(m.table.View())
			s.WriteString("\n\n")
			completed := len(filterTasks(m.tasks, func(t Task) bool { return t.Status == StatusDone }))
			inProgress := len(filterTasks(m.tasks, func(t Task) bool { return t.Status == StatusInProgress }))
			summary := fmt.Sprintf("  %d/%d tasks completed", completed, len(m.tasks))
			if inProgress > 0 {
				summary += fmt.Sprintf(", %d in progress", inProgress)
			}
			if m.filteredTasks != nil {
				summary += fmt.Sprintf(" (showing %d)", len(m.filteredTasks))
			}
			s.WriteString(infoStyle.Render(summary))
		}

		if m.mode == modeFilter {
			s.WriteString("\n\n")
			s.WriteString(statusStyle.Render(fmt.Sprintf("  / %s▌", m.filterText)))
		} else if m.status != "" {
			s.WriteString("\n\n")
			s.WriteString(statusStyle.Render("  " + m.status))
		}

		s.WriteString("\n\n")
		if m.mode == modeFilter {
			s.WriteString(helpStyle.Render("  type to filter · enter accept · esc clear"))
		} else if len(m.tasks) == 0 {
			help := "  a add · v view day · q quit"
			if m.latestDateWithTasks != "" {
				help = "  a add · i import · v view day · q quit"
			}
			s.WriteString(helpStyle.Render(help))
		} else {
			s.WriteString(helpStyle.Render("  a add · s start · d done · e/↵ edit · x delete · c carry · / search · 1-9 jump · v view · q quit"))
		}
		s.WriteString("\n")

	case modeConfirmCarry:
		toDate, _ := time.Parse("2006-01-02", tomorrow(m.date))
		s.WriteString(fmt.Sprintf("  Carry %d task(s) to %s:\n\n", len(m.carryCandidates), toDate.Format("02/01/2006")))
		for _, t := range m.carryCandidates {
			s.WriteString(fmt.Sprintf("    - %s\n", t.Description))
		}
		s.WriteString("\n")
		s.WriteString(m.form.View())

	default:
		s.WriteString(m.form.View())
	}

	return s.String()
}

// --- Filter mode ---

func (m *model) visibleTasks() []Task {
	if m.filteredTasks != nil {
		return m.filteredTasks
	}
	return m.tasks
}

func (m *model) applyFilter() {
	if m.filterText == "" {
		m.filteredTasks = nil
	} else {
		needle := strings.ToLower(m.filterText)
		m.filteredTasks = filterTasks(m.tasks, func(t Task) bool {
			return strings.Contains(strings.ToLower(t.Description), needle)
		})
	}
	m.rebuildTable()
}

func (m *model) clearFilter() {
	m.filterText = ""
	m.filteredTasks = nil
	m.rebuildTable()
}

func (m *model) updateFilter(msg tea.Msg) (tea.Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.clearFilter()
		m.mode = modeTable
		return m, nil
	case "enter":
		m.mode = modeTable
		return m, nil
	case "backspace":
		if len(m.filterText) > 0 {
			m.filterText = m.filterText[:len(m.filterText)-1]
			m.applyFilter()
		}
		return m, nil
	default:
		r := keyMsg.Runes
		if len(r) == 1 && r[0] >= 32 {
			m.filterText += string(r)
			m.applyFilter()
		}
		return m, nil
	}
}

// --- Table mode ---

func (m *model) updateTable(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		m.status = "" // clear status on any keypress
		switch keyMsg.String() {
		case "q":
			return m, tea.Quit
		case "a":
			return m.enterAddMode()
		case "s":
			return m.toggleInProgress()
		case "d":
			return m.toggleDone()
		case "e", "enter":
			return m.enterEditMode()
		case "i":
			return m.importTasks()
		case "x":
			return m.enterDeleteMode()
		case "c":
			return m.enterCarryMode()
		case "v":
			return m.enterViewDateMode()
		case "/":
			m.filterText = ""
			m.filteredTasks = nil
			m.mode = modeFilter
			m.rebuildTable()
			return m, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			num := int(keyMsg.Runes[0] - '0') // 1-based task number
			// Find the original task by its 1-based index in m.tasks
			if num <= len(m.tasks) {
				target := m.tasks[num-1]
				// Find its position in the visible (possibly filtered) list
				for vi, vt := range m.visibleTasks() {
					if vt.ID == target.ID {
						m.table.SetCursor(vi)
						break
					}
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *model) toggleDone() (tea.Model, tea.Cmd) {
	if len(m.tasks) == 0 {
		m.status = "No tasks."
		return m, nil
	}

	task := m.visibleTasks()[m.table.Cursor()]
	if task.Status == StatusDone {
		m.store.MarkIncomplete(task.ID)
		m.status = "Task marked as not done."
	} else {
		m.store.MarkComplete(task.ID)
		m.status = "Task marked as done."
	}

	m.refreshTasks()
	return m, nil
}

func (m *model) toggleInProgress() (tea.Model, tea.Cmd) {
	if len(m.tasks) == 0 {
		m.status = "No tasks."
		return m, nil
	}

	task := m.visibleTasks()[m.table.Cursor()]
	if task.Status == StatusInProgress {
		m.store.MarkIncomplete(task.ID)
		m.status = "Task no longer in progress."
	} else {
		m.store.MarkInProgress(task.ID)
		m.status = "Task marked as in progress."
	}

	m.refreshTasks()
	return m, nil
}

// --- Form modes ---

func (m *model) enterAddMode() (tea.Model, tea.Cmd) {
	m.formDesc = ""
	m.formPriority = PriorityB
	m.formEstimate = ""
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("What do you need to do?").Value(&m.formDesc).Validate(notEmpty("Description")),
			huh.NewSelect[Priority]().Title("Priority?").Options(PriorityOptions()...).Value(&m.formPriority),
			huh.NewInput().Title("Time estimate? (eg 30m, 2h, 1d)").Value(&m.formEstimate).Validate(notEmpty("Time estimate")),
		),
	)
	m.mode = modeAdd
	return m, m.form.Init()
}

func (m *model) enterEditMode() (tea.Model, tea.Cmd) {
	if len(m.tasks) == 0 {
		m.status = "No tasks to edit."
		return m, nil
	}

	task := m.visibleTasks()[m.table.Cursor()]
	m.editTaskID = task.ID
	m.formDesc = task.Description
	m.formPriority = task.Priority
	m.formEstimate = task.TimeEstimate
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Description").Value(&m.formDesc).Validate(notEmpty("Description")),
			huh.NewSelect[Priority]().Title("Priority?").Options(PriorityOptions()...).Value(&m.formPriority),
			huh.NewInput().Title("Time estimate?").Value(&m.formEstimate).Validate(notEmpty("Time estimate")),
		),
	)
	m.mode = modeEdit
	return m, m.form.Init()
}

func (m *model) enterDeleteMode() (tea.Model, tea.Cmd) {
	if len(m.tasks) == 0 {
		m.status = "No tasks to delete."
		return m, nil
	}

	task := m.visibleTasks()[m.table.Cursor()]
	m.editTaskID = task.ID
	m.formConfirm = true
	m.form = confirmForm(fmt.Sprintf("Delete '%s'?", task.Description), &m.formConfirm)
	m.mode = modeConfirmDelete
	return m, m.form.Init()
}

func (m *model) enterCarryMode() (tea.Model, tea.Cmd) {
	toDate := tomorrow(m.date)
	candidates, err := m.store.GetCarryOverCandidates(m.date, toDate, m.context)
	if err != nil {
		m.status = "Error loading tasks."
		return m, nil
	}

	if len(candidates) == 0 {
		incomplete := filterTasks(m.tasks, func(t Task) bool { return t.Status != StatusDone })
		if len(incomplete) == 0 {
			m.status = "No incomplete tasks to carry over."
		} else {
			m.status = "All incomplete tasks already carried over."
		}
		return m, nil
	}

	m.carryCandidates = candidates
	m.formConfirm = true
	m.form = confirmForm("Carry these over?", &m.formConfirm)
	m.mode = modeConfirmCarry
	return m, m.form.Init()
}

func (m *model) enterViewDateMode() (tea.Model, tea.Cmd) {
	m.formDate = ""
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Date (dd/mm/yyyy)").Value(&m.formDate).Validate(notEmpty("Date")),
		),
	)
	m.mode = modeViewDate
	return m, m.form.Init()
}

func (m *model) updateForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok && keyMsg.String() == "esc" {
		m.mode = modeTable
		m.status = ""
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	if m.form.State == huh.StateCompleted {
		return m.handleFormComplete()
	}

	return m, cmd
}

func (m *model) handleFormComplete() (tea.Model, tea.Cmd) {
	switch m.mode {
	case modeAdd:
		if err := m.store.AddTask(m.date, m.formDesc, m.formPriority, m.formEstimate, m.context); err != nil {
			m.status = "Error adding task."
		} else {
			m.status = "Task added."
		}

	case modeEdit:
		if err := m.store.UpdateTask(m.editTaskID, m.formDesc, m.formPriority, m.formEstimate); err != nil {
			m.status = "Error updating task."
		} else {
			m.status = "Task updated."
		}

	case modeConfirmDelete:
		if m.formConfirm {
			if err := m.store.DeleteTask(m.editTaskID); err != nil {
				m.status = "Error deleting task."
			} else {
				m.status = "Task deleted."
			}
		}

	case modeConfirmCarry:
		if m.formConfirm {
			toDate := tomorrow(m.date)
			if err := m.store.CarryOverTasks(m.carryCandidates, toDate, m.context); err != nil {
				m.status = "Error carrying over tasks."
			} else {
				m.status = "Tasks carried over."
			}
		}

	case modeViewDate:
		t, err := time.Parse("02/01/2006", m.formDate)
		if err != nil {
			m.status = "Invalid date. Use dd/mm/yyyy."
		} else {
			m.date = t.Format("2006-01-02")
			m.status = ""
		}
	}

	m.mode = modeTable
	m.refreshTasks()
	return m, nil
}

// --- Helpers ---

func (m *model) importTasks() (tea.Model, tea.Cmd) {
	if m.latestDateWithTasks == "" {
		return m, nil
	}

	if err := m.store.CopyIncompleteTasks(m.latestDateWithTasks, m.date, m.context); err != nil {
		m.status = "Error importing tasks."
	} else {
		lt, _ := time.Parse("2006-01-02", m.latestDateWithTasks)
		m.status = fmt.Sprintf("Tasks imported from %s.", lt.Format("02/01/2006"))
	}

	m.refreshTasks()
	return m, nil
}

func (m *model) refreshTasks() {
	tasks, err := m.store.GetTasksForDate(m.date, m.context)
	if err != nil {
		m.tasks = nil
	} else {
		m.tasks = tasks
	}

	m.latestDateWithTasks = ""
	if len(m.tasks) == 0 {
		if date, err := m.store.GetLatestDateWithIncompleteTasks(m.date, m.context); err == nil {
			m.latestDateWithTasks = date
		}
	}

	m.applyFilter()
}

func (m *model) rebuildTable() {
	cursor := m.table.Cursor()

	visible := m.visibleTasks()
	cols := tableColumns(m.width)
	rows := make([]table.Row, len(visible))
	for i, t := range visible {
		// Show the original 1-based index so number-jump stays consistent
		origIdx := i + 1
		for j, orig := range m.tasks {
			if orig.ID == t.ID {
				origIdx = j + 1
				break
			}
		}
		rows[i] = table.Row{
			fmt.Sprintf("%d", origIdx),
			t.DisplayDescription(),
			string(t.Priority),
			t.TimeEstimate,
			t.Status.Symbol(),
		}
	}

	height := len(visible) + 2 // +2 for header row + border
	if height < 3 {
		height = 3
	}
	if maxH := m.height - 10; maxH > 3 && height > maxH {
		height = maxH
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#fff")).
		Background(lipgloss.Color("#7c3aed")).
		Bold(true)
	t.SetStyles(s)

	// Preserve cursor position
	if cursor >= len(visible) {
		cursor = len(visible) - 1
	}
	if cursor < 0 {
		cursor = 0
	}
	t.SetCursor(cursor)

	m.table = t
}

func tableColumns(width int) []table.Column {
	fixed := 4 + 10 + 8 + 6 + 8 // #, Priority, Time, Status + padding/borders
	taskWidth := width - fixed
	if taskWidth < 20 {
		taskWidth = 20
	}
	if taskWidth > 80 {
		taskWidth = 80
	}
	return []table.Column{
		{Title: "#", Width: 4},
		{Title: "Task", Width: taskWidth},
		{Title: "Priority", Width: 10},
		{Title: "Time", Width: 8},
		{Title: "Status", Width: 6},
	}
}

func formatHeading(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.Format("Monday 2 January 2006")
}

func tomorrow(date string) string {
	t, _ := time.Parse("2006-01-02", date)
	return t.AddDate(0, 0, 1).Format("2006-01-02")
}

func confirmForm(title string, value *bool) *huh.Form {
	km := huh.NewDefaultKeyMap()
	km.Confirm.Toggle = key.NewBinding(key.WithKeys("h", "l", "right", "left", "tab"))
	return huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Affirmative("Yes").
				Negative("No").
				Value(value),
		),
	).WithKeyMap(km)
}

func notEmpty(field string) func(string) error {
	return func(s string) error {
		if s == "" {
			return fmt.Errorf("%s is required", field)
		}
		return nil
	}
}
