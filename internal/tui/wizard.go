package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// WizardPhase tracks which step the wizard is on.
type WizardPhase int

const (
	PhaseRepos  WizardPhase = iota // Add/manage repos
	PhaseBrowse                    // Browse skills.sh
	PhaseSkills                    // Enable/disable skills
	PhaseDone
)

var (
	wizardHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")).MarginBottom(1)
	stepStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	activeStepStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	inputStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errorStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	successStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

// WizardRepo is a repo entry shown in the wizard.
type WizardRepo struct {
	Name string
	URL  string
	New  bool // true if added during this wizard session
}

// WizardResult contains everything the wizard collected.
type WizardResult struct {
	NewRepos     []WizardRepo // repos added during the wizard
	BrowseNeeded bool         // user wants to browse skills.sh
	SkillChanges []SkillItem  // updated skill enable/disable states
	Cancelled    bool
}

// --- Phase 1: Repo management ---

type repoWizardModel struct {
	existingRepos []WizardRepo
	newRepos      []WizardRepo
	input         string
	errorMsg      string
	phase         WizardPhase
	done          bool
}

func newRepoWizardModel(existing []WizardRepo) repoWizardModel {
	return repoWizardModel{
		existingRepos: existing,
		phase:         PhaseRepos,
	}
}

func (m repoWizardModel) Init() tea.Cmd { return nil }

func (m repoWizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.done = true
			m.phase = PhaseDone
			return m, tea.Quit
		case "enter":
			if m.input == "" {
				// No input — move to next phase
				m.done = true
				return m, tea.Quit
			}
			if m.input == "b" || m.input == "browse" {
				m.phase = PhaseBrowse
				m.done = true
				return m, tea.Quit
			}
			// Validate input looks like owner/repo or URL
			input := strings.TrimSpace(m.input)
			if input == "" {
				m.input = ""
				return m, nil
			}
			m.newRepos = append(m.newRepos, WizardRepo{
				Name: input,
				URL:  input,
				New:  true,
			})
			m.input = ""
			m.errorMsg = ""
		case "backspace", "ctrl+h":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "ctrl+u":
			m.input = ""
		default:
			if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
				m.input += msg.String()
				m.errorMsg = ""
			}
		}
	}
	return m, nil
}

func (m repoWizardModel) View() string {
	var b strings.Builder

	// Step indicator
	b.WriteString(renderSteps(0))
	b.WriteString("\n\n")

	b.WriteString(wizardHeaderStyle.Render("Step 1: Manage Repos"))
	b.WriteString("\n\n")

	// Show existing repos
	if len(m.existingRepos) > 0 || len(m.newRepos) > 0 {
		b.WriteString("  Current repos:\n")
		for _, r := range m.existingRepos {
			b.WriteString(fmt.Sprintf("    %s %s\n", successStyle.Render("●"), r.Name))
		}
		for _, r := range m.newRepos {
			b.WriteString(fmt.Sprintf("    %s %s %s\n", successStyle.Render("●"), r.Name, successStyle.Render("(new)")))
		}
		b.WriteString("\n")
	}

	// Input prompt
	b.WriteString("  Add a repo: ")
	b.WriteString(inputStyle.Render(m.input))
	b.WriteString(inputStyle.Render("█"))
	b.WriteString("\n")

	if m.errorMsg != "" {
		b.WriteString("  " + errorStyle.Render(m.errorMsg) + "\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  type owner/repo or URL  'b' to browse skills.sh  enter to continue"))
	return b.String()
}

func renderSteps(active int) string {
	steps := []string{"Repos", "Skills", "Sync"}
	var parts []string
	for i, s := range steps {
		if i == active {
			parts = append(parts, activeStepStyle.Render(fmt.Sprintf("(%d) %s", i+1, s)))
		} else if i < active {
			parts = append(parts, successStyle.Render(fmt.Sprintf("✓ %s", s)))
		} else {
			parts = append(parts, stepStyle.Render(fmt.Sprintf("(%d) %s", i+1, s)))
		}
	}
	return "  " + strings.Join(parts, "  →  ")
}

// RunRepoWizard shows the repo management step.
// Returns new repos to add and whether to browse skills.sh.
func RunRepoWizard(existing []WizardRepo) (newRepos []WizardRepo, wantBrowse bool, err error) {
	m := newRepoWizardModel(existing)
	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, false, err
	}
	rm := result.(repoWizardModel)
	return rm.newRepos, rm.phase == PhaseBrowse, nil
}
