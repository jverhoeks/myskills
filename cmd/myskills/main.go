package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/jverhoeks/myskills/internal/config"
	"github.com/jverhoeks/myskills/internal/detect"
	"github.com/jverhoeks/myskills/internal/dev"
	gh "github.com/jverhoeks/myskills/internal/github"
	"github.com/jverhoeks/myskills/internal/manifest"
	"github.com/jverhoeks/myskills/internal/repo"
	"github.com/jverhoeks/myskills/internal/skill"
	"github.com/jverhoeks/myskills/internal/sync"
	"github.com/jverhoeks/myskills/internal/tui"
	"github.com/jverhoeks/myskills/internal/validate"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "myskills",
		Short:   "Distribute AI agent skills from a central repo",
		Version: version,
	}

	root.AddCommand(
		newInitCmd(),
		newAddRepoCmd(),
		newAddSkillCmd(),
		newSyncCmd(),
		newListCmd(),
		newInfoCmd(),
		newValidateCmd(),
		newEnableCmd(),
		newSearchCmd(),
		newDevCmd(),
		newSubmitCmd(),
		newRemoveCmd(),
		newDoctorCmd(),
		newConfigCmd(),
	)

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadConfig() (config.Config, error) {
	return config.Load(config.Path())
}

func enabledTargets(cfg config.Config) map[string]string {
	targets := make(map[string]string)
	for name, t := range cfg.Targets {
		if t.Enabled {
			targets[name] = config.ExpandPath(t.SkillPath)
		}
	}
	return targets
}

func manifestPath() string {
	return filepath.Join(config.Dir(), "manifest.json")
}

func targetNames(targets map[string]string) []string {
	var names []string
	for name := range targets {
		names = append(names, name)
	}
	return names
}

// repoSkillFilter returns a filter that checks if a skill from a given repo is enabled.
func repoSkillFilter(cfg config.Config, repoName string) func(string) bool {
	return func(name string) bool {
		return cfg.IsSkillEnabledInRepo(repoName, name)
	}
}

// --- init ---

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init <repo-url|owner/repo> [name]",
		Short: "Set up myskills with a repo, detect tools, write config",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoURL := repo.ResolveURL(args[0])
			repoName := repo.NameFromURL(args[0])
			if len(args) == 2 {
				repoName = args[1]
			}

			cfg := config.Default()
			cfg.Repos = []config.Repo{{Name: repoName, URL: repoURL}}

			cacheDir := config.RepoDir(repoName)
			fmt.Printf("Cloning %s to %s...\n", repoURL, cacheDir)
			if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
				return fmt.Errorf("creating cache dir: %w", err)
			}
			if err := repo.Clone(repoURL, cacheDir); err != nil {
				return err
			}
			fmt.Println("  ✓ Cloned")

			fmt.Println("\nDetected AI tools:")
			detected := detect.Detect()
			for _, name := range []string{"claude", "copilot", "codex", "opencode"} {
				marker := "✗"
				if detected[name] {
					marker = "✓"
				}
				fmt.Printf("  %s %s\n", marker, name)
			}

			fmt.Println()
			for _, name := range []string{"claude", "copilot", "codex", "opencode"} {
				if !detected[name] {
					continue
				}
				fmt.Printf("Enable %s? [Y/n] ", name)
				var answer string
				fmt.Scanln(&answer)
				answer = strings.ToLower(strings.TrimSpace(answer))
				t := cfg.Targets[name]
				t.Enabled = answer == "" || answer == "y" || answer == "yes"
				cfg.Targets[name] = t
			}

			// Auto-enable all skills from the initial repo
			cacheDir = config.RepoDir(repoName)
			skills, _ := repo.ListSkills(cacheDir)
			if len(skills) > 0 {
				cfg.EnableSkills(repoName, skills)
				fmt.Printf("\n  Enabled %d skill(s): %s\n", len(skills), strings.Join(skills, ", "))
			}

			cfgPath := config.Path()
			if err := config.Save(cfg, cfgPath); err != nil {
				return err
			}
			fmt.Printf("\nConfig written to %s\n", cfgPath)
			fmt.Println("Run 'myskills sync' to install skills.")
			return nil
		},
	}
}

// --- add-repo ---

func newAddRepoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-repo <url|owner/repo> [name]",
		Short: "Add another skill repository",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}

			repoURL := repo.ResolveURL(args[0])
			repoName := repo.NameFromURL(args[0])
			if len(args) == 2 {
				repoName = args[1]
			}

			// Check for duplicate names
			for _, r := range cfg.Repos {
				if r.Name == repoName {
					return fmt.Errorf("repo %q already exists", repoName)
				}
			}

			cacheDir := config.RepoDir(repoName)
			fmt.Printf("Cloning %s to %s...\n", repoURL, cacheDir)
			if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
				return fmt.Errorf("creating cache dir: %w", err)
			}
			if err := repo.Clone(repoURL, cacheDir); err != nil {
				return err
			}
			fmt.Println("  ✓ Cloned")

			cfg.Repos = append(cfg.Repos, config.Repo{Name: repoName, URL: repoURL})

			// List available skills but don't auto-enable
			skills, _ := repo.ListSkills(cacheDir)
			if err := config.Save(cfg, config.Path()); err != nil {
				return err
			}
			fmt.Printf("✓ Added repo %q (%d skills available)\n", repoName, len(skills))
			fmt.Println("Run 'myskills enable' to choose which skills to activate.")
			return nil
		},
	}
}

// --- add-skill (skills.sh compatible) ---

func newAddSkillCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add-skill <owner/repo>",
		Short: "Add a skill from skills.sh / GitHub (owner/repo shorthand)",
		Long: `Add a skill from any GitHub repository using owner/repo shorthand.
Compatible with skills.sh — same repos that work with "npx skills add" work here.

Examples:
  myskills add-skill vercel-labs/agent-skills
  myskills add-skill anthropics/courses
  myskills add-skill microsoft/skills`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}

			input := args[0]
			repoURL := repo.ResolveURL(input)
			repoName := repo.NameFromURL(input)

			// Check for duplicate
			for _, r := range cfg.Repos {
				if r.Name == repoName {
					return fmt.Errorf("repo %q already exists — run 'myskills sync' to update it", repoName)
				}
			}

			cacheDir := config.RepoDir(repoName)
			fmt.Printf("Adding %s...\n", input)
			if err := os.MkdirAll(filepath.Dir(cacheDir), 0o755); err != nil {
				return fmt.Errorf("creating cache dir: %w", err)
			}
			if err := repo.Clone(repoURL, cacheDir); err != nil {
				return err
			}

			// Discover skills
			skills, _ := repo.ListSkills(cacheDir)
			if len(skills) == 0 {
				fmt.Printf("  ⚠ No skills found in %s\n", input)
				fmt.Println("  Checked: skills/, .agents/skills/, .claude/skills/, .github/skills/, and root SKILL.md")
				return nil
			}

			cfg.Repos = append(cfg.Repos, config.Repo{Name: repoName, URL: repoURL})
			if err := config.Save(cfg, config.Path()); err != nil {
				return err
			}

			fmt.Printf("  ✓ Found %d skill(s): %s\n", len(skills), strings.Join(skills, ", "))
			fmt.Printf("  ✓ Added repo %q\n", repoName)
			fmt.Println("\nSkills are disabled by default. Run 'myskills enable' to choose which ones to activate.")
			return nil
		},
	}
}

// --- sync ---

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [skill-name]",
		Short: "Pull latest and symlink skills to tool directories",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}
			if len(cfg.Repos) == 0 {
				return fmt.Errorf("no repos configured — run 'myskills init'")
			}

			targets := enabledTargets(cfg)
			if len(targets) == 0 {
				return fmt.Errorf("no targets enabled — run 'myskills config set targets.<name>.enabled true'")
			}

			totalCount := 0
			for _, r := range cfg.Repos {
				cacheDir := config.RepoDir(r.Name)
				fmt.Printf("[%s] Pulling latest...\n", r.Name)
				if err := repo.Pull(cacheDir); err != nil {
					fmt.Printf("[%s] ✗ pull failed: %v\n", r.Name, err)
					continue
				}

				if len(args) == 1 {
					name := args[0]
					if !cfg.IsSkillEnabledInRepo(r.Name, name) {
						fmt.Printf("[%s] %s is disabled — run 'myskills enable'\n", r.Name, name)
						continue
					}
					skillDir := repo.SkillDir(cacheDir, name)
					if err := sync.One(skillDir, name, targets); err != nil {
						continue
					}
					fmt.Printf("[%s] ✓ %s linked to %s\n", r.Name, name, strings.Join(targetNames(targets), ", "))
					totalCount++
				} else {
					// Build skill map: name → source path
					skillNames, _ := repo.ListSkills(cacheDir)
					skillMap := make(map[string]string, len(skillNames))
					for _, name := range skillNames {
						skillMap[name] = repo.SkillDir(cacheDir, name)
					}
					count, err := sync.AllFromMap(skillMap, targets, repoSkillFilter(cfg, r.Name))
					if err != nil {
						fmt.Printf("[%s] ✗ sync failed: %v\n", r.Name, err)
						continue
					}
					fmt.Printf("[%s] ✓ %d skills linked to %s\n", r.Name, count, strings.Join(targetNames(targets), ", "))
					totalCount += count
				}
			}

			if len(args) == 0 {
				fmt.Printf("\n✓ %d total skills linked\n", totalCount)
			}

			// Update manifest
			mPath := manifestPath()
			m, _ := manifest.Load(mPath)
			m.LastSync = time.Now().UTC()
			for _, r := range cfg.Repos {
				cacheDir := config.RepoDir(r.Name)
				skills, _ := repo.ListSkills(cacheDir)
				for _, name := range skills {
					hash, _ := repo.CommitHashForPath(cacheDir, repo.SkillDir(cacheDir, name))
					m.Skills[name] = manifest.SkillEntry{
						Commit:   hash,
						SyncedAt: time.Now().UTC(),
					}
				}
			}
			manifest.Save(m, mPath)

			return nil
		},
	}
}

// --- list ---

func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List skills with enabled/synced status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}

			m, _ := manifest.Load(manifestPath())

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "REPO\tSKILL\tENABLED\tSTATUS\tSYNCED")

			for _, r := range cfg.Repos {
				cacheDir := config.RepoDir(r.Name)
				skills, err := repo.ListSkills(cacheDir)
				if err != nil {
					continue
				}
				for _, name := range skills {
					hash, _ := repo.CommitHashForPath(cacheDir, repo.SkillDir(cacheDir, name))
					enabled := "yes"
					if !cfg.IsSkillEnabledInRepo(r.Name, name) {
						enabled = "no"
					}
					status := "not installed"
					synced := "-"
					if entry, ok := m.Skills[name]; ok {
						synced = entry.SyncedAt.Format("2006-01-02 15:04")
						if entry.Commit == hash {
							status = "current"
						} else {
							status = "outdated"
						}
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", r.Name, name, enabled, status, synced)
				}
			}
			w.Flush()
			return nil
		},
	}
}

// --- info ---

func newInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show skill details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			name := args[0]

			// Find skill across repos
			for _, r := range cfg.Repos {
				cacheDir := config.RepoDir(r.Name)
				skillDir := repo.SkillDir(cacheDir, name)
				skillPath := filepath.Join(skillDir, "SKILL.md")

				s, err := skill.Parse(skillPath)
				if err != nil {
					continue
				}

				fmt.Printf("Name:        %s\n", s.Name)
				fmt.Printf("Repo:        %s\n", r.Name)
				fmt.Printf("Description: %s\n", s.Description)
				if team := s.Metadata["team"]; team != "" {
					fmt.Printf("Team:        %s\n", team)
				}
				enabled := "yes"
				if !cfg.IsSkillEnabledInRepo(r.Name, name) {
					enabled = "no"
				}
				fmt.Printf("Enabled:     %s\n", enabled)

				fmt.Println("\nFiles:")
				filepath.WalkDir(skillDir, func(path string, d os.DirEntry, err error) error {
					if err != nil {
						return nil
					}
					rel, _ := filepath.Rel(skillDir, path)
					if rel == "." {
						return nil
					}
					fmt.Printf("  %s\n", rel)
					return nil
				})
				return nil
			}

			return fmt.Errorf("skill %q not found in any repo", name)
		},
	}
}

// --- enable ---

func newEnableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "enable",
		Short: "Toggle which skills are enabled (interactive picker)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}

			var items []tui.SkillItem
			for _, r := range cfg.Repos {
				cacheDir := config.RepoDir(r.Name)
				skills, err := repo.ListSkills(cacheDir)
				if err != nil {
					continue
				}
				for _, name := range skills {
					skillPath := filepath.Join(repo.SkillDir(cacheDir, name), "SKILL.md")
					desc := ""
					if s, err := skill.Parse(skillPath); err == nil {
						desc = s.Description
					}

					qualified := r.Name + ":" + name

					items = append(items, tui.SkillItem{
						Name:        qualified,
						Description: desc,
						Enabled:     cfg.IsSkillEnabledInRepo(r.Name, name),
					})
				}
			}

			if len(items) == 0 {
				fmt.Println("No skills found in any repo.")
				return nil
			}

			result, err := tui.RunPicker(items)
			if err != nil {
				return err
			}

			// Update config — item.Name is already "repo:skill" qualified
			for _, item := range result {
				cfg.SetSkillEnabled(item.Name, item.Enabled)
			}

			if err := config.Save(cfg, config.Path()); err != nil {
				return err
			}

			enabled := 0
			for _, item := range result {
				if item.Enabled {
					enabled++
				}
			}
			fmt.Printf("✓ %d/%d skills enabled. Run 'myskills sync' to apply.\n", enabled, len(result))
			return nil
		},
	}
}

// --- search ---

func newSearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search skills by name or description",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}

			query := strings.ToLower(strings.Join(args, " "))
			found := 0

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "REPO\tSKILL\tENABLED\tDESCRIPTION")

			for _, r := range cfg.Repos {
				cacheDir := config.RepoDir(r.Name)
				skills, err := repo.ListSkills(cacheDir)
				if err != nil {
					continue
				}
				for _, name := range skills {
					skillPath := filepath.Join(repo.SkillDir(cacheDir, name), "SKILL.md")
					desc := ""
					if s, parseErr := skill.Parse(skillPath); parseErr == nil {
						desc = s.Description
					}

					// Match against name and description
					if !strings.Contains(strings.ToLower(name), query) &&
						!strings.Contains(strings.ToLower(desc), query) {
						continue
					}

					enabled := "yes"
					if !cfg.IsSkillEnabledInRepo(r.Name, name) {
						enabled = "no"
					}

					// Truncate description for table
					short := desc
					if len(short) > 70 {
						short = short[:67] + "..."
					}

					fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", r.Name, name, enabled, short)
					found++
				}
			}
			w.Flush()

			if found == 0 {
				fmt.Printf("No skills matching %q\n", query)
			} else {
				fmt.Printf("\n%d result(s)\n", found)
			}
			return nil
		},
	}
}

// --- validate ---

func newValidateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate <path>",
		Short: "Validate a skill against spec and org rules",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			errs := validate.Spec(path)

			// Try to load org rules from each cached repo
			cfg, cfgErr := loadConfig()
			if cfgErr == nil {
				for _, r := range cfg.Repos {
					rulesPath := filepath.Join(config.RepoDir(r.Name), ".myskills.yaml")
					if rules, err := validate.LoadOrgRules(rulesPath); err == nil {
						errs = append(errs, validate.Org(path, rules)...)
						break // Use first repo's rules
					}
				}
			}

			if len(errs) == 0 {
				fmt.Printf("✓ %s is valid\n", path)
				return nil
			}

			fmt.Printf("✗ %s has %d issue(s):\n", path, len(errs))
			for _, e := range errs {
				fmt.Printf("  - %s\n", e)
			}
			return fmt.Errorf("validation failed")
		},
	}
}

// --- dev ---

func newDevCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dev <name>",
		Short: "Scaffold a new skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			devDir := filepath.Join(config.Dir(), "dev")
			name := args[0]

			if err := dev.Scaffold(devDir, name); err != nil {
				return err
			}

			skillDir := filepath.Join(devDir, name)
			fmt.Printf("Created skill scaffold at %s\n", skillDir)
			fmt.Println("\nEdit SKILL.md, then run:")
			fmt.Printf("  myskills validate %s\n", skillDir)
			fmt.Printf("  myskills submit %s\n", name)
			return nil
		},
	}
}

// --- submit ---

func newSubmitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit <name>",
		Short: "Validate and open a PR for a skill",
		Args:  cobra.ExactArgs(1),
	}
	cmd.Flags().String("repo", "", "Target repo name (defaults to first repo)")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return fmt.Errorf("load config: %w", err)
		}
		if len(cfg.Repos) == 0 {
			return fmt.Errorf("no repos configured")
		}

		name := args[0]
		devDir := filepath.Join(config.Dir(), "dev", name)

		fmt.Println("Validating...")
		errs := validate.Spec(devDir)
		if len(errs) > 0 {
			for _, e := range errs {
				fmt.Printf("  ✗ %s\n", e)
			}
			return fmt.Errorf("validation failed — fix issues before submitting")
		}
		fmt.Println("  ✓ Valid")

		// Pick target repo
		repoName, _ := cmd.Flags().GetString("repo")
		targetRepo := cfg.Repos[0]
		if repoName != "" {
			found := false
			for _, r := range cfg.Repos {
				if r.Name == repoName {
					targetRepo = r
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("repo %q not found", repoName)
			}
		}

		cacheDir := config.RepoDir(targetRepo.Name)
		dst := repo.SkillDir(cacheDir, name)
		fmt.Printf("Copying to repo %q...\n", targetRepo.Name)
		if err := sync.CopySkill(devDir, dst); err != nil {
			return err
		}
		fmt.Println("  ✓ Copied")

		fmt.Println("Creating PR...")
		if err := gh.Submit(cacheDir, name, cfg.GitHub.Method); err != nil {
			return err
		}

		return nil
	}
	return cmd
}

// --- remove ---

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a skill from all tool directories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
			}

			name := args[0]
			targets := enabledTargets(cfg)
			removed := 0

			for tName, tPath := range targets {
				skillPath := filepath.Join(tPath, name)
				if _, err := os.Lstat(skillPath); err == nil {
					if err := sync.RemoveSkill(skillPath); err != nil {
						return fmt.Errorf("removing from %s: %w", tName, err)
					}
					removed++
				}
			}

			if removed == 0 {
				fmt.Printf("Skill %q not found in any target\n", name)
			} else {
				fmt.Printf("✓ Removed %s from %d target(s)\n", name, removed)
			}

			mPath := manifestPath()
			m, _ := manifest.Load(mPath)
			delete(m.Skills, name)
			manifest.Save(m, mPath)

			return nil
		},
	}
}

// --- doctor ---

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check health: repos, tools, config",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Checking health...")
			fmt.Println()

			cfg, err := loadConfig()
			if err != nil {
				fmt.Println("Config:  ✗ " + err.Error())
				fmt.Println("\nRun 'myskills init <url>' to set up.")
				return nil
			}
			fmt.Printf("Config:  ✓ %s\n", config.Path())
			fmt.Printf("Cache:   %s\n", config.CacheDir())

			fmt.Println("\nRepos:")
			for _, r := range cfg.Repos {
				cacheDir := config.RepoDir(r.Name)
				if _, err := os.Stat(cacheDir); err != nil {
					fmt.Printf("  %s: ✗ not cloned (%s)\n", r.Name, r.URL)
				} else {
					skills, _ := repo.ListSkills(cacheDir)
					fmt.Printf("  %s: ✓ %d skills (%s)\n", r.Name, len(skills), r.URL)
				}
			}

			fmt.Println("\nTargets:")
			for name, t := range cfg.Targets {
				if !t.Enabled {
					fmt.Printf("  %s: disabled\n", name)
					continue
				}
				path := config.ExpandPath(t.SkillPath)
				if _, err := os.Stat(path); err != nil {
					fmt.Printf("  %s: ✓ enabled (dir will be created on sync)\n", name)
				} else {
					entries, _ := os.ReadDir(path)
					symlinks := 0
					for _, e := range entries {
						info, err := os.Lstat(filepath.Join(path, e.Name()))
						if err == nil && info.Mode()&os.ModeSymlink != 0 {
							symlinks++
						}
					}
					fmt.Printf("  %s: ✓ enabled (%d skills linked)\n", name, symlinks)
				}
			}

			if len(cfg.Skills.Enabled) > 0 {
				fmt.Printf("\nEnabled skills: %s\n", strings.Join(cfg.Skills.Enabled, ", "))
			} else {
				fmt.Println("\nEnabled skills: (none)")
			}

			fmt.Println()
			if gh.HasGH() {
				fmt.Println("gh CLI:      ✓ available")
			} else {
				fmt.Println("gh CLI:      ✗ not found")
			}
			if gh.HasToken() {
				fmt.Println("GITHUB_TOKEN: ✓ set")
			} else {
				fmt.Println("GITHUB_TOKEN: not set")
			}

			return nil
		},
	}
}

// --- config ---

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Show current config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			fmt.Println("repos:")
			for _, r := range cfg.Repos {
				fmt.Printf("  - %s: %s\n", r.Name, r.URL)
			}
			fmt.Printf("cache_dir: %s\n", config.CacheDir())
			fmt.Printf("github.method: %s\n", cfg.GitHub.Method)
			fmt.Println("\ntargets:")
			for name, t := range cfg.Targets {
				fmt.Printf("  %s: enabled=%v path=%s\n", name, t.Enabled, t.SkillPath)
			}
			if len(cfg.Skills.Enabled) > 0 {
				fmt.Printf("\nenabled skills: %s\n", strings.Join(cfg.Skills.Enabled, ", "))
			} else {
				fmt.Println("\nenabled skills: (none — run 'myskills enable' to activate)")
			}
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return err
			}

			key, value := args[0], args[1]

			switch key {
			case "github.method":
				cfg.GitHub.Method = value
			default:
				parts := strings.SplitN(key, ".", 3)
				if len(parts) == 3 && parts[0] == "targets" {
					name := parts[1]
					field := parts[2]
					t, ok := cfg.Targets[name]
					if !ok {
						return fmt.Errorf("unknown target: %s", name)
					}
					switch field {
					case "enabled":
						t.Enabled = value == "true"
					case "skill_path":
						t.SkillPath = value
					default:
						return fmt.Errorf("unknown target field: %s", field)
					}
					cfg.Targets[name] = t
				} else {
					return fmt.Errorf("unknown config key: %s", key)
				}
			}

			if err := config.Save(cfg, config.Path()); err != nil {
				return err
			}
			fmt.Printf("Set %s = %s\n", key, value)
			return nil
		},
	})

	return cmd
}
