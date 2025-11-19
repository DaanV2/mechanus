package tui

import (
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var _ tea.Model = &Table[bool]{}
var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

// Table is a generic table widget that displays items of type T as rows.
type Table[T any] struct {
	table table.Model
	conv  func(item T) []string
}

// NewTable creates a new table with the provided converter function for items.
func NewTable[T any](conv func(item T) []string) *Table[T] {
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)

	t := table.New(
		table.WithFocused(true),
		table.WithHeight(7),
		table.WithStyles(s),
	)
	t.Help.ShowAll = true

	return &Table[T]{
		conv:  conv,
		table: t,
	}
}

func (t *Table[T]) SetColumns(c ...table.Column) {
	t.table.SetColumns(c)
}

func (t *Table[T]) AddItem(item T) {
	data := t.conv(item)
	rows := t.table.Rows()
	rows = append(rows, data)
	t.table.SetRows(rows)
}

func (t *Table[T]) AddItems(items []T) {
	rows := t.table.Rows()

	for _, item := range items {
		data := t.conv(item)
		rows = append(rows, data)
	}

	t.table.SetRows(rows)
}

// Init implements tea.Model.
func (t *Table[T]) Init() tea.Cmd {
	return nil
}

func (t *Table[T]) AutoWidth() {
	cols := t.table.Columns()
	rows := t.table.Rows()

	for _, r := range rows {
		length := min(len(cols), len(r))
		for c := range length {
			cols[c].Width = max(len(r[c]), cols[c].Width)
		}
	}

	t.table.SetColumns(cols)
}

// Update implements tea.Model.
func (t *Table[T]) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if kmsg, ok := msg.(tea.KeyMsg); ok {
		switch kmsg.Type { //nolint:exhaustive //we dont need to do all types here, just the ones we want
		case tea.KeyCtrlC, tea.KeyCtrlD:
			return t, tea.Quit
		}

		if kmsg.String() == "q" {
			return t, tea.Quit
		}
	}

	m, c := t.table.Update(msg)
	t.table = m

	return t, c
}

// View implements tea.Model.
func (t *Table[T]) View() string {
	return baseStyle.Render(t.table.View()) + "\n  " + t.table.HelpView() + "\n"
}
