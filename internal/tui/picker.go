package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	checkedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	uncheckedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan
	titleStyle     = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// SkillItem represents a skill in the picker.
type SkillItem struct {
	Name        string
	Description string
	Enabled     bool
}

type model struct {
	items    []SkillItem
	cursor   int
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "enter":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ", "x":
			m.items[m.cursor].Enabled = !m.items[m.cursor].Enabled
		case "a":
			// Toggle all
			allEnabled := true
			for _, item := range m.items {
				if !item.Enabled {
					allEnabled = false
					break
				}
			}
			for i := range m.items {
				m.items[i].Enabled = !allEnabled
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("Select skills to enable"))
	b.WriteString("\n")

	for i, item := range m.items {
		cursor := "  "
		if m.cursor == i {
			cursor = cursorStyle.Render("> ")
		}

		check := uncheckedStyle.Render("[ ]")
		if item.Enabled {
			check = checkedStyle.Render("[x]")
		}

		name := item.Name
		if m.cursor == i {
			name = cursorStyle.Render(name)
		}

		desc := ""
		if item.Description != "" {
			// Truncate long descriptions
			d := item.Description
			if len(d) > 60 {
				d = d[:57] + "..."
			}
			desc = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(" — " + d)
		}

		fmt.Fprintf(&b, "%s%s %s%s\n", cursor, check, name, desc)
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("space: toggle  a: toggle all  enter/q: save  esc: cancel"))
	return b.String()
}

// RunPicker shows an interactive checkbox list and returns the updated items.
func RunPicker(items []SkillItem) ([]SkillItem, error) {
	m := model{items: items}
	p := tea.NewProgram(m)
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running picker: %w", err)
	}
	return result.(model).items, nil
}
