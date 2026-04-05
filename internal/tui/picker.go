package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultPageSize = 20

var (
	checkedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	uncheckedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))  // gray
	cursorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan
	titleStyle     = lipgloss.NewStyle().Bold(true)
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	filterStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow
	countStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

// SkillItem represents a skill in the picker.
type SkillItem struct {
	Name        string
	Description string
	Enabled     bool
}

type model struct {
	allItems []SkillItem // all items (source of truth)
	filtered []int       // indices into allItems that match the filter
	cursor   int         // position within filtered list
	offset   int         // scroll offset for viewport
	pageSize int         // visible items per page
	filter   string      // current search filter
	quitting bool
}

func newModel(items []SkillItem) model {
	m := model{
		allItems: items,
		pageSize: defaultPageSize,
	}
	m.applyFilter()
	return m
}

func (m *model) applyFilter() {
	m.filtered = m.filtered[:0]
	query := strings.ToLower(m.filter)
	for i, item := range m.allItems {
		if query == "" ||
			strings.Contains(strings.ToLower(item.Name), query) ||
			strings.Contains(strings.ToLower(item.Description), query) {
			m.filtered = append(m.filtered, i)
		}
	}
	// Reset cursor if it's out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.clampOffset()
}

func (m *model) clampOffset() {
	if len(m.filtered) <= m.pageSize {
		m.offset = 0
		return
	}
	// Ensure cursor is visible
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.pageSize {
		m.offset = m.cursor - m.pageSize + 1
	}
	// Clamp offset
	if m.offset < 0 {
		m.offset = 0
	}
	maxOffset := len(m.filtered) - m.pageSize
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust page size to terminal height (leave room for header/footer)
		available := msg.Height - 6
		if available < 5 {
			available = 5
		}
		m.pageSize = available
		m.clampOffset()

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "ctrl+p":
			if m.cursor > 0 {
				m.cursor--
				m.clampOffset()
			}
		case "down", "ctrl+n":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.clampOffset()
			}
		case "pgup":
			m.cursor -= m.pageSize
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.clampOffset()
		case "pgdown":
			m.cursor += m.pageSize
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			m.clampOffset()
		case "home":
			m.cursor = 0
			m.clampOffset()
		case "end":
			m.cursor = len(m.filtered) - 1
			m.clampOffset()
		case " ":
			if len(m.filtered) > 0 {
				idx := m.filtered[m.cursor]
				m.allItems[idx].Enabled = !m.allItems[idx].Enabled
			}
		case "ctrl+a":
			// Toggle all visible (filtered) items
			allEnabled := true
			for _, idx := range m.filtered {
				if !m.allItems[idx].Enabled {
					allEnabled = false
					break
				}
			}
			for _, idx := range m.filtered {
				m.allItems[idx].Enabled = !allEnabled
			}
		case "backspace", "ctrl+h":
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
				m.applyFilter()
			}
		case "ctrl+u":
			m.filter = ""
			m.applyFilter()
		default:
			// Type to filter — only single printable chars
			if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
				m.filter += msg.String()
				m.applyFilter()
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

	// Header
	enabledCount := 0
	for _, item := range m.allItems {
		if item.Enabled {
			enabledCount++
		}
	}
	header := fmt.Sprintf("Select skills to enable (%d/%d enabled)", enabledCount, len(m.allItems))
	b.WriteString(titleStyle.Render(header))
	b.WriteString("\n")

	// Filter line
	if m.filter != "" {
		b.WriteString(filterStyle.Render(fmt.Sprintf("  filter: %s", m.filter)))
		b.WriteString(countStyle.Render(fmt.Sprintf("  (%d matching)", len(m.filtered))))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(m.filtered) == 0 {
		if m.filter != "" {
			b.WriteString("  No skills matching filter\n")
		} else {
			b.WriteString("  No skills found\n")
		}
	} else {
		// Scroll indicator top
		if m.offset > 0 {
			b.WriteString(countStyle.Render(fmt.Sprintf("  ↑ %d more above\n", m.offset)))
		}

		// Visible window
		end := m.offset + m.pageSize
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		for vi := m.offset; vi < end; vi++ {
			idx := m.filtered[vi]
			item := m.allItems[idx]

			cursor := "  "
			if vi == m.cursor {
				cursor = cursorStyle.Render("> ")
			}

			check := uncheckedStyle.Render("[ ]")
			if item.Enabled {
				check = checkedStyle.Render("[x]")
			}

			name := item.Name
			if vi == m.cursor {
				name = cursorStyle.Render(name)
			}

			desc := ""
			if item.Description != "" {
				d := item.Description
				if len(d) > 55 {
					d = d[:52] + "..."
				}
				desc = countStyle.Render(" — " + d)
			}

			fmt.Fprintf(&b, "%s%s %s%s\n", cursor, check, name, desc)
		}

		// Scroll indicator bottom
		remaining := len(m.filtered) - end
		if remaining > 0 {
			b.WriteString(countStyle.Render(fmt.Sprintf("  ↓ %d more below\n", remaining)))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("space: toggle  ctrl+a: toggle all  type: filter  ctrl+u: clear filter"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("pgup/pgdn: page  home/end: jump  enter: save  esc: cancel"))
	return b.String()
}

// RunPicker shows an interactive checkbox list and returns the updated items.
func RunPicker(items []SkillItem) ([]SkillItem, error) {
	m := newModel(items)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running picker: %w", err)
	}
	return result.(model).allItems, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
