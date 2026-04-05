package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// RepoItem represents a repo in the browse picker.
type RepoItem struct {
	OwnerRepo  string // "owner/repo"
	SkillCount int
	Selected   bool
	AlreadyAdded bool
}

type repoModel struct {
	allItems []RepoItem
	filtered []int
	cursor   int
	offset   int
	pageSize int
	filter   string
	quitting bool
}

func newRepoModel(items []RepoItem) repoModel {
	m := repoModel{
		allItems: items,
		pageSize: defaultPageSize,
	}
	m.applyFilter()
	return m
}

func (m *repoModel) applyFilter() {
	m.filtered = m.filtered[:0]
	query := strings.ToLower(m.filter)
	for i, item := range m.allItems {
		if query == "" || strings.Contains(strings.ToLower(item.OwnerRepo), query) {
			m.filtered = append(m.filtered, i)
		}
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.clampOffset()
}

func (m *repoModel) clampOffset() {
	if len(m.filtered) <= m.pageSize {
		m.offset = 0
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.pageSize {
		m.offset = m.cursor - m.pageSize + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	maxOff := len(m.filtered) - m.pageSize
	if m.offset > maxOff {
		m.offset = maxOff
	}
}

func (m repoModel) Init() tea.Cmd { return nil }

func (m repoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
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
			// Cancel — deselect everything
			for i := range m.allItems {
				m.allItems[i].Selected = false
			}
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
				if !m.allItems[idx].AlreadyAdded {
					m.allItems[idx].Selected = !m.allItems[idx].Selected
				}
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
			if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
				m.filter += msg.String()
				m.applyFilter()
			}
		}
	}
	return m, nil
}

var (
	addedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
	repoCountStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

func (m repoModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	selected := 0
	for _, item := range m.allItems {
		if item.Selected {
			selected++
		}
	}
	header := fmt.Sprintf("Browse skills.sh repos (%d selected)", selected)
	b.WriteString(titleStyle.Render(header))
	b.WriteString("\n")

	if m.filter != "" {
		b.WriteString(filterStyle.Render(fmt.Sprintf("  filter: %s", m.filter)))
		b.WriteString(countStyle.Render(fmt.Sprintf("  (%d matching)", len(m.filtered))))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	if len(m.filtered) == 0 {
		b.WriteString("  No repos matching filter\n")
	} else {
		if m.offset > 0 {
			b.WriteString(countStyle.Render(fmt.Sprintf("  ↑ %d more above\n", m.offset)))
		}

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

			if item.AlreadyAdded {
				name := addedStyle.Render(item.OwnerRepo + " (already added)")
				fmt.Fprintf(&b, "%s    %s\n", cursor, name)
				continue
			}

			check := uncheckedStyle.Render("[ ]")
			if item.Selected {
				check = checkedStyle.Render("[x]")
			}

			name := item.OwnerRepo
			if vi == m.cursor {
				name = cursorStyle.Render(name)
			}

			skills := ""
			if item.SkillCount > 0 {
				skills = repoCountStyle.Render(fmt.Sprintf(" (%d skills)", item.SkillCount))
			}

			fmt.Fprintf(&b, "%s%s %s%s\n", cursor, check, name, skills)
		}

		remaining := len(m.filtered) - end
		if remaining > 0 {
			b.WriteString(countStyle.Render(fmt.Sprintf("  ↓ %d more below\n", remaining)))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("space: select  type: filter  ctrl+u: clear  enter: add selected  esc: cancel"))
	return b.String()
}

// RunRepoPicker shows an interactive repo picker and returns selected repos.
func RunRepoPicker(items []RepoItem) ([]RepoItem, error) {
	m := newRepoModel(items)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running picker: %w", err)
	}

	var selected []RepoItem
	for _, item := range result.(repoModel).allItems {
		if item.Selected {
			selected = append(selected, item)
		}
	}
	return selected, nil
}
