package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/sbp/myskills/internal/config"
	"github.com/sbp/myskills/internal/detect"
	"github.com/sbp/myskills/internal/dev"
	gh "github.com/sbp/myskills/internal/github"
	"github.com/sbp/myskills/internal/manifest"
	"github.com/sbp/myskills/internal/repo"
	"github.com/sbp/myskills/internal/skill"
	"github.com/sbp/myskills/internal/sync"
	"github.com/sbp/myskills/internal/validate"

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
		newSyncCmd(),
		newListCmd(),
		newInfoCmd(),
		newValidateCmd(),
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

// --- init ---

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Configure repo URL, detect tools, write config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Default()

			fmt.Print("Repository URL: ")
			var repoURL string
			fmt.Scanln(&repoURL)
			if repoURL == "" {
				return fmt.Errorf("repository URL is required")
			}
			cfg.Repo = repoURL

			cacheDir := config.ExpandPath(cfg.CacheDir)
			fmt.Printf("Cloning repository to %s...\n", cacheDir)
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

// --- sync ---

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [skill-name]",
		Short: "Pull latest and copy skills to tool directories",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}

			cacheDir := config.ExpandPath(cfg.CacheDir)
			fmt.Println("Pulling latest...")
			if err := repo.Pull(cacheDir); err != nil {
				return err
			}

			targets := enabledTargets(cfg)
			if len(targets) == 0 {
				return fmt.Errorf("no targets enabled — run 'myskills config set targets.<name>.enabled true'")
			}

			if len(args) == 1 {
				name := args[0]
				if err := sync.One(cacheDir, name, targets); err != nil {
					return err
				}
				fmt.Printf("✓ %s synced to %s\n", name, strings.Join(targetNames(targets), ", "))
			} else {
				count, err := sync.All(cacheDir, targets)
				if err != nil {
					return err
				}
				fmt.Printf("✓ %d skills synced to %s\n", count, strings.Join(targetNames(targets), ", "))
			}

			// Update manifest
			mPath := manifestPath()
			m, _ := manifest.Load(mPath)
			m.LastSync = time.Now().UTC()

			skills, _ := repo.ListSkills(cacheDir)
			for _, name := range skills {
				hash, _ := repo.CommitHashForPath(cacheDir, filepath.Join("skills", name))
				m.Skills[name] = manifest.SkillEntry{
					Commit:   hash,
					SyncedAt: time.Now().UTC(),
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
		Short: "List installed skills with status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w (run 'myskills init' first)", err)
			}

			cacheDir := config.ExpandPath(cfg.CacheDir)
			skills, err := repo.ListSkills(cacheDir)
			if err != nil {
				return err
			}

			m, _ := manifest.Load(manifestPath())

			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "SKILL\tSTATUS\tSYNCED")
			for _, name := range skills {
				hash, _ := repo.CommitHashForPath(cacheDir, filepath.Join("skills", name))
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
				fmt.Fprintf(w, "%s\t%s\t%s\n", name, status, synced)
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

			cacheDir := config.ExpandPath(cfg.CacheDir)
			skillDir := repo.SkillDir(cacheDir, args[0])
			skillPath := filepath.Join(skillDir, "SKILL.md")

			s, err := skill.Parse(skillPath)
			if err != nil {
				return err
			}

			fmt.Printf("Name:        %s\n", s.Name)
			fmt.Printf("Description: %s\n", s.Description)
			if team := s.Metadata["team"]; team != "" {
				fmt.Printf("Team:        %s\n", team)
			}

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

			// Try to load org rules from cached repo
			cfg, cfgErr := loadConfig()
			if cfgErr == nil {
				cacheDir := config.ExpandPath(cfg.CacheDir)
				rulesPath := filepath.Join(cacheDir, ".myskills.yaml")
				if rules, err := validate.LoadOrgRules(rulesPath); err == nil {
					errs = append(errs, validate.Org(path, rules)...)
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
	return &cobra.Command{
		Use:   "submit <name>",
		Short: "Validate and open a PR for a skill",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := loadConfig()
			if err != nil {
				return fmt.Errorf("load config: %w", err)
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

			cacheDir := config.ExpandPath(cfg.CacheDir)
			dst := repo.SkillDir(cacheDir, name)
			fmt.Println("Copying to repo...")
			if err := sync.CopySkill(devDir, dst); err != nil {
				return err
			}
			fmt.Println("  ✓ Copied")

			fmt.Println("Creating PR...")
			if err := gh.Submit(cacheDir, name, cfg.GitHub.Method, cfg.GitHub.Token); err != nil {
				return err
			}

			return nil
		},
	}
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
				skillDir := filepath.Join(tPath, name)
				if _, err := os.Stat(skillDir); err == nil {
					if err := sync.RemoveSkill(skillDir); err != nil {
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
		Short: "Check health: repo, tools, config",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("Checking health...")
			fmt.Println()

			cfg, err := loadConfig()
			if err != nil {
				fmt.Println("Config:  ✗ " + err.Error())
				fmt.Println("\nRun 'myskills init' to set up.")
				return nil
			}
			fmt.Printf("Config:  ✓ %s\n", config.Path())
			fmt.Printf("Repo:    %s\n", cfg.Repo)

			cacheDir := config.ExpandPath(cfg.CacheDir)
			if _, err := os.Stat(cacheDir); err != nil {
				fmt.Println("Cache:   ✗ repo not cloned")
			} else {
				skills, _ := repo.ListSkills(cacheDir)
				fmt.Printf("Cache:   ✓ %s (%d skills)\n", cacheDir, len(skills))
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
					fmt.Printf("  %s: ✓ enabled (%d skills installed)\n", name, len(entries))
				}
			}

			if gh.HasGH() {
				fmt.Println("\ngh CLI: ✓ available")
			} else {
				fmt.Println("\ngh CLI: ✗ not found (submit will use token or manual mode)")
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

			fmt.Printf("repo: %s\n", cfg.Repo)
			fmt.Printf("cache_dir: %s\n", cfg.CacheDir)
			fmt.Printf("github.method: %s\n", cfg.GitHub.Method)
			if cfg.GitHub.Token != "" {
				fmt.Println("github.token: [set]")
			}
			fmt.Println("\ntargets:")
			for name, t := range cfg.Targets {
				fmt.Printf("  %s: enabled=%v path=%s\n", name, t.Enabled, t.SkillPath)
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
			case "repo":
				cfg.Repo = value
			case "cache_dir":
				cfg.CacheDir = value
			case "github.method":
				cfg.GitHub.Method = value
			case "github.token":
				cfg.GitHub.Token = value
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
